package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
	"time"

	"github.com/shinerio/skillflow/core/applog"
	"github.com/shinerio/skillflow/core/backup"
	"github.com/shinerio/skillflow/core/config"
	coregit "github.com/shinerio/skillflow/core/git"
	"github.com/shinerio/skillflow/core/install"
	"github.com/shinerio/skillflow/core/notify"
	"github.com/shinerio/skillflow/core/registry"
	"github.com/shinerio/skillflow/core/skill"
	"github.com/shinerio/skillflow/core/skillkey"
	agentsync "github.com/shinerio/skillflow/core/sync"
	"github.com/shinerio/skillflow/core/upgrade"
	"github.com/shinerio/skillflow/core/viewstate"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type maxDepthPuller interface {
	PullWithMaxDepth(ctx context.Context, sourceDir string, maxDepth int) ([]*skill.Skill, error)
}

type githubSkillDownloader interface {
	DownloadTo(ctx context.Context, source install.InstallSource, c install.SkillCandidate, targetDir string) error
}

type App struct {
	ctx                context.Context
	hub                *notify.Hub
	sysLog             *applog.Logger
	storage            *skill.Storage
	config             *config.Service
	starStorage        *coregit.StarStorage
	cacheDir           string
	viewCache          *viewstate.Manager
	startupOnce        sync.Once
	initialWindowState config.WindowState
	autostartFactory   func() (launchAtLoginController, error)
	ghDownloader       func(client *http.Client) githubSkillDownloader

	// Git sync state
	gitConflictMu      sync.Mutex
	gitConflictPending bool
	stopAutoSync       chan struct{}

	backupResultMu       sync.RWMutex
	lastBackupResult     []backup.RemoteFile
	lastBackupAt         time.Time
	windowVisibilityMu   sync.Mutex
	windowVisibilityInit bool
	windowVisible        bool
	uiControlMu          sync.Mutex
	uiControlServer      *loopbackControlServer
}

const defaultCategoryName = "Default"

var builtinStarredRepoURLs = []string{
	"https://github.com/anthropics/skills.git",
	"https://github.com/ComposioHQ/awesome-claude-skills.git",
	"https://github.com/affaan-m/everything-claude-code.git",
}

var appDataDirFunc = config.AppDataDir

var runStartupUpgrade = upgrade.Run

var loadStartupConfig = func(dataDir string) (*config.Service, config.AppConfig, error) {
	svc := config.NewService(dataDir)
	cfg, err := svc.Load()
	return svc, cfg, err
}

func normalizeCategoryName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || strings.EqualFold(trimmed, defaultCategoryName) {
		return defaultCategoryName
	}
	return trimmed
}

func NewApp() *App {
	return &App{
		hub: notify.NewHub(),
		ghDownloader: func(client *http.Client) githubSkillDownloader {
			return install.NewGitHubInstaller("", client)
		},
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	dataDir := appDataDirFunc()
	if ctx != nil {
		runtime.LogInfof(ctx, "config upgrade started: dataDir=%s", dataDir)
	}
	if err := runStartupUpgrade(dataDir); err != nil {
		if ctx != nil {
			runtime.LogErrorf(ctx, "config upgrade failed: dataDir=%s err=%v", dataDir, err)
		}
		return
	}
	if ctx != nil {
		runtime.LogInfof(ctx, "config upgrade completed: dataDir=%s", dataDir)
	}

	var cfg config.AppConfig
	var err error
	a.config, cfg, err = loadStartupConfig(dataDir)
	if err != nil {
		runtime.LogErrorf(ctx, "load config failed: %v", err)
		cfg = config.DefaultConfig(dataDir)
	}
	a.initLogger(cfg.LogLevel)
	a.logInfof("application startup, version=%s, dataDir=%s", Version, dataDir)
	if err := a.syncLaunchAtLogin(cfg.LaunchAtLogin); err != nil {
		a.logErrorf("application startup launch-at-login reconcile failed: %v", err)
	}
	a.storage = skill.NewStorage(cfg.SkillsStorageDir)
	a.cacheDir = filepath.Join(dataDir, "cache")
	a.viewCache = viewstate.NewManager(filepath.Join(a.cacheDir, "viewstate"))
	a.starStorage = coregit.NewStarStorageWithBuiltins(filepath.Join(dataDir, "star_repos.json"), builtinStarredRepoURLs)
	registerAdapters()
	registerProviders()
	if ctx != nil {
		go forwardEvents(ctx, a.hub)
	}
	a.logDebugf("startup background tasks deferred until ui ready")
	a.startAutoSyncTimer(cfg.Cloud.SyncIntervalMinutes)
}

// proxyHTTPClient builds an *http.Client configured according to the saved proxy settings.
// Falls back to http.DefaultClient on any error.
func (a *App) proxyHTTPClient() *http.Client {
	cfg, err := a.config.Load()
	if err != nil {
		return http.DefaultClient
	}
	switch cfg.Proxy.Mode {
	case config.ProxyModeSystem:
		return &http.Client{Transport: &http.Transport{Proxy: http.ProxyFromEnvironment}}
	case config.ProxyModeManual:
		if cfg.Proxy.URL == "" {
			return http.DefaultClient
		}
		proxyURL, err := url.Parse(cfg.Proxy.URL)
		if err != nil {
			return http.DefaultClient
		}
		return &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)}}
	default: // ProxyModeNone or empty
		return &http.Client{Transport: &http.Transport{Proxy: nil}}
	}
}

func (a *App) domReady(ctx context.Context) {
	a.fitInitialWindowToScreen(ctx)
	a.setupTrayForUI(ctx)
	a.startUIControlServer()
	a.startBackgroundStartupTasks()
}

func (a *App) startBackgroundStartupTasks() {
	a.startupOnce.Do(func() {
		tasks := a.startupBackgroundTaskPlan()
		a.logDebugf("startup background tasks scheduled, count=%d", len(tasks))
		scheduleStartupBackgroundTasks(tasks, func(task startupBackgroundTask) {
			time.AfterFunc(task.delay, func() {
				a.logDebugf("startup background task started: task=%s delay=%s", task.name, task.delay)
				task.run()
			})
		})
	})
}

func (a *App) beforeClose(ctx context.Context) bool {
	a.persistCurrentWindowSize(ctx)
	a.publishWindowVisibilityChanged(false)
	a.logInfof("application quit started")
	return false
}

func (a *App) shutdown(_ context.Context) {
	a.stopUIControlServer()
	a.teardownTrayForUI()
	a.logInfof("application shutdown completed")
}

func (a *App) showMainWindow() {
	a.logInfof("main window show started")
	if err := showMainWindowNative(a.ctx); err != nil {
		a.logErrorf("main window show failed: %v", err)
		return
	}
	a.publishWindowVisibilityChanged(true)
	a.logInfof("main window show completed")
}

func (a *App) hideMainWindow() {
	if goruntime.GOOS != "darwin" {
		a.logInfof("main window hide started")
	}
	if err := hideMainWindowNative(a.ctx); err != nil {
		a.logErrorf("main window hide failed: %v", err)
		return
	}
	a.publishWindowVisibilityChanged(false)
	if goruntime.GOOS != "darwin" {
		a.logInfof("main window hide completed")
	}
}

func (a *App) quitApp() {
	runtime.Quit(a.ctx)
}

// autoBackup triggers cloud backup after any mutating operation if cloud is enabled.
func (a *App) autoBackup() {
	cfg, ok := a.loadAutoBackupConfig()
	if !ok {
		return
	}
	a.autoBackupWithConfig(cfg)
}

func (a *App) loadAutoBackupConfig() (config.AppConfig, bool) {
	if a.config == nil {
		return config.AppConfig{}, false
	}
	cfg, err := a.config.Load()
	if err != nil {
		a.logErrorf("auto backup skipped: load config failed: %v", err)
		return config.AppConfig{}, false
	}
	if !cfg.Cloud.Enabled || cfg.Cloud.Provider == "" {
		return config.AppConfig{}, false
	}
	return cfg, true
}

func (a *App) autoBackupWithConfig(cfg config.AppConfig) {
	a.logDebugf("auto backup triggered")
	_ = a.runBackupWithConfig(cfg)
}

func (a *App) scheduleAutoBackup() {
	cfg, err := a.config.Load()
	if err != nil {
		a.logErrorf("auto backup scheduling failed: load config failed: %v", err)
		return
	}
	if !cfg.Cloud.Enabled || strings.TrimSpace(cfg.Cloud.Provider) == "" {
		a.logDebugf("auto backup scheduling skipped: reason=cloud-disabled")
		return
	}
	go a.autoBackup()
}

func (a *App) runBackup() error {
	cfg, err := a.config.Load()
	if err != nil || !cfg.Cloud.Enabled || cfg.Cloud.Provider == "" {
		if err != nil {
			a.logErrorf("backup aborted: load config failed: %v", err)
		}
		return err
	}
	return a.runBackupWithConfig(cfg)
}

func (a *App) runBackupWithConfig(cfg config.AppConfig) error {
	a.logInfof("backup started (provider=%s)", cfg.Cloud.Provider)
	provider, ok := registry.GetCloudProvider(cfg.Cloud.Provider)
	if !ok {
		return fmt.Errorf("provider not found: %s", cfg.Cloud.Provider)
	}
	if err := provider.Init(cfg.Cloud.Credentials); err != nil {
		return err
	}
	isGit := cfg.Cloud.Provider == backup.GitProviderName
	backupDir := a.backupRootDir(cfg)
	var err error
	if isGit {
		backupDir, err = a.prepareGitBackupRoot(cfg)
		if err != nil {
			return err
		}
	}
	changes, currentSnapshot, err := a.computeBackupChanges(provider, backupDir, isGit)
	if err != nil {
		a.logErrorf("backup failed: compute changes failed: %v", err)
		return err
	}
	if isGit {
		a.hub.Publish(notify.Event{Type: notify.EventGitSyncStarted})
	}
	a.hub.Publish(notify.Event{Type: notify.EventBackupStarted})
	err = provider.Sync(a.ctx, backupDir, cfg.Cloud.BucketName, cfg.Cloud.RemotePath,
		func(file string) {
			a.hub.Publish(notify.Event{
				Type:    notify.EventBackupProgress,
				Payload: notify.BackupProgressPayload{CurrentFile: file},
			})
		})
	if err != nil {
		a.logErrorf("backup failed: %v", err)
		var conflictErr *backup.GitConflictError
		if isGit && errors.As(err, &conflictErr) {
			a.publishGitConflict(conflictErr)
		}
		if isGit {
			a.hub.Publish(notify.Event{Type: notify.EventGitSyncFailed, Payload: err.Error()})
		}
		a.hub.Publish(notify.Event{Type: notify.EventBackupFailed, Payload: err.Error()})
		return err
	} else {
		a.recordBackupResult(changes, currentSnapshot)
		payload := notify.BackupCompletedPayload{Files: changes, CompletedAt: a.GetLastBackupCompletedAt()}
		a.logInfof("backup completed")
		if isGit {
			a.clearGitConflictPending()
			a.hub.Publish(notify.Event{Type: notify.EventGitSyncCompleted, Payload: payload})
		}
		a.hub.Publish(notify.Event{Type: notify.EventBackupCompleted, Payload: payload})
		return nil
	}
}

func (a *App) publishGitConflict(conflictErr *backup.GitConflictError) {
	a.gitConflictMu.Lock()
	a.gitConflictPending = true
	a.gitConflictMu.Unlock()
	a.hub.Publish(notify.Event{
		Type: notify.EventGitConflict,
		Payload: notify.GitConflictPayload{
			Message: conflictErr.Output,
			Files:   conflictErr.Files,
		},
	})
}

func (a *App) clearGitConflictPending() {
	a.gitConflictMu.Lock()
	a.gitConflictPending = false
	a.gitConflictMu.Unlock()
}

func (a *App) reloadStateFromDisk() {
	cfg, err := a.config.Load()
	if err != nil {
		return
	}
	a.storage = skill.NewStorage(cfg.SkillsStorageDir)
	a.startAutoSyncTimer(cfg.Cloud.SyncIntervalMinutes)
}

// gitPullOnStartup pulls from the remote git repo at startup when the git provider is enabled.
func (a *App) gitPullOnStartup() {
	cfg, err := a.config.Load()
	if err != nil || !cfg.Cloud.Enabled || cfg.Cloud.Provider != backup.GitProviderName {
		return
	}
	beforeRestore := a.captureCloudRestoreState()
	a.logInfof("startup git pull started")
	p, ok := registry.GetCloudProvider(backup.GitProviderName)
	if !ok {
		return
	}
	if err := p.Init(cfg.Cloud.Credentials); err != nil {
		return
	}
	gitP := p.(*backup.GitProvider)
	backupDir, prepErr := a.prepareGitBackupRoot(cfg)
	if prepErr != nil {
		a.logErrorf("startup git pull failed: prepare git backup root failed: %v", prepErr)
		a.hub.Publish(notify.Event{Type: notify.EventGitSyncFailed, Payload: prepErr.Error()})
		return
	}
	beforeSnapshot, snapErr := backup.BuildSnapshot(backupDir)
	if snapErr != nil {
		a.logErrorf("startup git pull failed: build pre-pull snapshot failed: %v", snapErr)
		a.hub.Publish(notify.Event{Type: notify.EventGitSyncFailed, Payload: snapErr.Error()})
		return
	}
	a.hub.Publish(notify.Event{Type: notify.EventGitSyncStarted})
	if err := gitP.Restore(a.ctx, "", "", backupDir); err != nil {
		a.logErrorf("startup git pull failed: %v", err)
		var conflictErr *backup.GitConflictError
		if errors.As(err, &conflictErr) {
			a.publishGitConflict(conflictErr)
			a.hub.Publish(notify.Event{Type: notify.EventGitSyncFailed, Payload: err.Error()})
		} else {
			a.hub.Publish(notify.Event{Type: notify.EventGitSyncFailed, Payload: err.Error()})
		}
		return
	}
	changes, afterSnapshot, diffErr := a.computeSnapshotChanges(beforeSnapshot, backupDir)
	if diffErr != nil {
		a.logErrorf("startup git pull failed: compute post-pull changes failed: %v", diffErr)
		a.hub.Publish(notify.Event{Type: notify.EventGitSyncFailed, Payload: diffErr.Error()})
		return
	}
	a.recordBackupResult(changes, afterSnapshot)
	if err := a.handleRestoredCloudState(beforeRestore, "startup.git.pull"); err != nil {
		a.logErrorf("startup git pull failed: restore compensation failed: %v", err)
		a.hub.Publish(notify.Event{Type: notify.EventGitSyncFailed, Payload: err.Error()})
		return
	}
	a.logInfof("startup git pull completed")
	a.clearGitConflictPending()
	a.hub.Publish(notify.Event{Type: notify.EventGitSyncCompleted, Payload: notify.BackupCompletedPayload{Files: changes, CompletedAt: a.GetLastBackupCompletedAt()}})
}

// startAutoSyncTimer starts (or restarts) a periodic auto-backup ticker.
// intervalMinutes <= 0 disables the timer.
func (a *App) startAutoSyncTimer(intervalMinutes int) {
	if a.stopAutoSync != nil {
		close(a.stopAutoSync)
		a.stopAutoSync = nil
	}
	if intervalMinutes <= 0 {
		return
	}
	stop := make(chan struct{})
	a.stopAutoSync = stop
	go func() {
		ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				a.autoBackup()
			case <-stop:
				return
			}
		}
	}()
}

// GetGitConflictPending returns true when a git conflict from startup pull is waiting to be resolved.
func (a *App) GetGitConflictPending() bool {
	a.gitConflictMu.Lock()
	defer a.gitConflictMu.Unlock()
	return a.gitConflictPending
}

// ResolveGitConflict resolves a pending git merge conflict.
// useLocal=true  → keep local changes, force-push to remote.
// useLocal=false → discard local changes, reset to remote state.
func (a *App) ResolveGitConflict(useLocal bool) error {
	action := "use_remote"
	if useLocal {
		action = "use_local"
	}
	p, ok := registry.GetCloudProvider(backup.GitProviderName)
	if !ok {
		return fmt.Errorf("git provider 未注册")
	}
	gitP := p.(*backup.GitProvider)
	localCfg := a.config.LoadLocalRuntimeConfig()
	backupDir := a.backupRootDir(config.AppConfig{SkillsStorageDir: localCfg.SkillsStorageDir})
	a.logInfof("git conflict resolution started: strategy=%s, backupDir=%s", action, backupDir)
	beforeSnapshot, snapErr := backup.BuildSnapshot(backupDir)
	if snapErr != nil {
		a.logErrorf("git conflict resolution failed: build pre-resolution snapshot failed: strategy=%s, backupDir=%s, err=%v", action, backupDir, snapErr)
		return snapErr
	}
	var err error
	if useLocal {
		err = gitP.ResolveConflictUseLocal(backupDir)
	} else {
		err = gitP.ResolveConflictUseRemote(backupDir)
	}
	if err != nil {
		a.logErrorf("git conflict resolution failed: strategy=%s, backupDir=%s, err=%v", action, backupDir, err)
		return err
	}
	a.gitConflictMu.Lock()
	a.gitConflictPending = false
	a.gitConflictMu.Unlock()
	changes, afterSnapshot, diffErr := a.computeSnapshotChanges(beforeSnapshot, backupDir)
	if diffErr != nil {
		a.logErrorf("git conflict resolution failed: compute post-resolution changes failed: strategy=%s, backupDir=%s, err=%v", action, backupDir, diffErr)
		return diffErr
	}
	a.recordBackupResult(changes, afterSnapshot)
	a.reloadStateFromDisk()
	a.hub.Publish(notify.Event{Type: notify.EventGitSyncCompleted, Payload: notify.BackupCompletedPayload{Files: changes, CompletedAt: a.GetLastBackupCompletedAt()}})
	a.logInfof("git conflict resolution completed: strategy=%s, backupDir=%s", action, backupDir)
	return nil
}

func (a *App) backupRootDir(cfg config.AppConfig) string {
	appDataDir := filepath.Clean(config.AppDataDir())
	skillsDir := filepath.Clean(cfg.SkillsStorageDir)

	// Prefer app data dir so cloud backup includes config/meta/skills.
	if rel, err := filepath.Rel(appDataDir, skillsDir); err == nil &&
		rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return appDataDir
	}
	// If skills dir is outside app data dir, use its parent as the git root.
	return filepath.Dir(skillsDir)
}

func (a *App) prepareGitBackupRoot(cfg config.AppConfig) (string, error) {
	backupDir := a.backupRootDir(cfg)
	migratedTo, migrated, err := backup.MigrateLegacyNestedGitDir(cfg.SkillsStorageDir, backupDir)
	if err != nil {
		a.logErrorf("git backup root preparation failed: skillsDir=%s, backupDir=%s, err=%v", cfg.SkillsStorageDir, backupDir, err)
		return "", err
	}
	if migrated {
		a.logInfof("git backup root preparation completed: moved legacy nested git dir from skills storage to %s", migratedTo)
	}
	return backupDir, nil
}

func (a *App) backupSnapshotPath() string {
	return filepath.Join(config.AppDataDir(), "cache", "backup_snapshot.json")
}

func (a *App) computeBackupChanges(provider backup.CloudProvider, backupDir string, isGit bool) ([]backup.RemoteFile, backup.Snapshot, error) {
	if isGit {
		if gitProvider, ok := provider.(*backup.GitProvider); ok {
			changes, err := gitProvider.PendingChanges(backupDir)
			return changes, nil, err
		}
	}

	previousSnapshot, err := backup.LoadSnapshot(a.backupSnapshotPath())
	if err != nil {
		return nil, nil, err
	}
	currentSnapshot, err := backup.BuildSnapshot(backupDir)
	if err != nil {
		return nil, nil, err
	}
	return backup.DiffSnapshots(previousSnapshot, currentSnapshot), currentSnapshot, nil
}

func (a *App) computeSnapshotChanges(beforeSnapshot backup.Snapshot, root string) ([]backup.RemoteFile, backup.Snapshot, error) {
	afterSnapshot, err := backup.BuildSnapshot(root)
	if err != nil {
		return nil, nil, err
	}
	return backup.DiffSnapshots(beforeSnapshot, afterSnapshot), afterSnapshot, nil
}

func (a *App) recordBackupResult(files []backup.RemoteFile, snapshot backup.Snapshot) {
	if snapshot != nil {
		if err := backup.SaveSnapshot(a.backupSnapshotPath(), snapshot); err != nil {
			a.logErrorf("backup snapshot save failed: %v", err)
		}
	}
	a.setLastBackupResult(files)
}

func (a *App) setLastBackupResult(files []backup.RemoteFile) {
	copied := make([]backup.RemoteFile, len(files))
	copy(copied, files)

	a.backupResultMu.Lock()
	a.lastBackupResult = copied
	a.lastBackupAt = time.Now().UTC()
	a.backupResultMu.Unlock()
}

func (a *App) GetLastBackupChanges() []backup.RemoteFile {
	a.backupResultMu.RLock()
	defer a.backupResultMu.RUnlock()

	copied := make([]backup.RemoteFile, len(a.lastBackupResult))
	copy(copied, a.lastBackupResult)
	return copied
}

func (a *App) GetLastBackupCompletedAt() string {
	a.backupResultMu.RLock()
	defer a.backupResultMu.RUnlock()

	if a.lastBackupAt.IsZero() {
		return ""
	}
	return a.lastBackupAt.Format(time.RFC3339)
}

func (a *App) gitProxyURL() string {
	cfg, err := a.config.Load()
	if err != nil {
		return ""
	}
	if cfg.Proxy.Mode == config.ProxyModeManual {
		return cfg.Proxy.URL
	}
	return ""
}

// --- Skills ---

func (a *App) ListSkills() ([]InstalledSkillEntry, error) {
	return measureOperation(a, "list_skills", func() ([]InstalledSkillEntry, error) {
		fingerprint, fingerprintErr := a.installedSkillsFingerprint()
		if fingerprintErr == nil {
			var cached []InstalledSkillEntry
			state, err := a.ensureViewCache().Load(installedSkillsSnapshotName, fingerprint, &cached)
			if err == nil && state == viewstate.StateHit {
				return cached, nil
			}
		}

		entries, err := a.listSkillsUncached()
		if err != nil {
			return nil, err
		}
		if fingerprint, err := a.installedSkillsFingerprint(); err == nil {
			_ = a.ensureViewCache().Save(installedSkillsSnapshotName, fingerprint, entries)
		}
		return entries, nil
	})
}

func (a *App) ListCategories() ([]string, error) {
	cats, err := a.storage.ListCategories()
	if err != nil {
		return nil, err
	}
	// 检查默认分类是否已在列表中
	hasDefault := false
	for _, c := range cats {
		if normalizeCategoryName(c) == defaultCategoryName {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		// 将默认分类加到列表最前面
		cats = append([]string{defaultCategoryName}, cats...)
	}
	return cats, nil
}

func (a *App) CreateCategory(name string) error {
	return a.storage.CreateCategory(name)
}

func (a *App) RenameCategory(oldName, newName string) error {
	if normalizeCategoryName(oldName) == defaultCategoryName {
		return fmt.Errorf("默认分类不可重命名")
	}
	if normalizeCategoryName(newName) == defaultCategoryName {
		return fmt.Errorf("不能重命名为默认分类")
	}
	return a.storage.RenameCategory(strings.TrimSpace(oldName), strings.TrimSpace(newName))
}

func (a *App) DeleteCategory(name string) error {
	name = strings.TrimSpace(name)
	a.logInfof("delete category started: category=%s", name)
	if normalizeCategoryName(name) == defaultCategoryName {
		err := fmt.Errorf("默认分类不可删除")
		a.logErrorf("delete category failed: category=%s, err=%v", name, err)
		return err
	}
	if err := a.storage.DeleteCategory(name); err != nil {
		if errors.Is(err, skill.ErrCategoryNotEmpty) {
			wrapped := fmt.Errorf("分类下仍有 Skill，请先清空后再删除")
			a.logErrorf("delete category failed: category=%s, err=%v", name, wrapped)
			return wrapped
		}
		a.logErrorf("delete category failed: category=%s, err=%v", name, err)
		return err
	}
	a.logInfof("delete category completed: category=%s", name)
	return nil
}

func (a *App) MoveSkillCategory(skillID, category string) error {
	category = normalizeCategoryName(category)
	return a.storage.MoveCategory(skillID, category)
}

func (a *App) DeleteSkill(skillID string) error {
	if err := a.storage.Delete(skillID); err != nil {
		return err
	}
	a.scheduleAutoBackup()
	return nil
}

func (a *App) DeleteSkills(skillIDs []string) error {
	for _, id := range skillIDs {
		if err := a.storage.Delete(id); err != nil {
			return err
		}
	}
	a.scheduleAutoBackup()
	return nil
}

func (a *App) GetSkillMeta(skillID string) (*skill.SkillMeta, error) {
	sk, err := a.storage.Get(skillID)
	if err != nil {
		return nil, err
	}
	return skill.ReadMeta(sk.Path)
}

// GetSkillMetaByPath reads skill.md frontmatter from a skill directory path (no ID required).
func (a *App) GetSkillMetaByPath(path string) (*skill.SkillMeta, error) {
	return skill.ReadMeta(path)
}

// ReadSkillFileContent returns the full text content of skill.md inside the given skill directory.
func (a *App) ReadSkillFileContent(path string) (string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if !e.IsDir() && strings.ToLower(e.Name()) == "skill.md" {
			data, err := os.ReadFile(filepath.Join(path, e.Name()))
			if err != nil {
				return "", err
			}
			return string(data), nil
		}
	}
	return "", fmt.Errorf("skill.md not found in %s", path)
}

// OpenURL opens the given URL in the system default browser.
// Non-HTTP URLs (e.g. SSH git remotes) are first converted to HTTPS.
func (a *App) OpenURL(rawURL string) error {
	target := rawURL
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		if canonical, err := coregit.CanonicalRepoURL(rawURL); err == nil {
			target = canonical
		}
	}
	runtime.BrowserOpenURL(a.ctx, target)
	return nil
}

// PushStarSkillsToAgents copies starred skill directories directly to the push directory of each
// specified agent, skipping skills that already exist. Returns structured conflicts.
func (a *App) PushStarSkillsToAgents(skillPaths []string, agentNames []string) ([]PushConflict, error) {
	a.logInfof("push starred skills to agents started: skillCount=%d agentCount=%d", len(skillPaths), len(agentNames))
	cfg, _ := a.config.Load()
	var conflicts []PushConflict
	for _, agentName := range agentNames {
		for _, agent := range cfg.Agents {
			if agent.Name != agentName {
				continue
			}
			if agent.PushDir == "" {
				err := fmt.Errorf("智能体 %s 未配置推送路径", agentName)
				a.logErrorf("push starred skills to agents failed: agent=%s err=%v", agentName, err)
				return nil, err
			}
			if err := os.MkdirAll(agent.PushDir, 0755); err != nil {
				a.logErrorf("push starred skills to agents failed: agent=%s pushDir=%s err=%v", agentName, agent.PushDir, err)
				return nil, err
			}
			adapter := getAdapter(agent)
			for _, skillPath := range skillPaths {
				name := filepath.Base(skillPath)
				dst := filepath.Join(agent.PushDir, name)
				if _, err := os.Stat(dst); err == nil {
					conflicts = append(conflicts, PushConflict{
						SkillName:  name,
						SkillPath:  skillPath,
						AgentName:  agentName,
						TargetPath: dst,
					})
					continue
				}
				sk := []*skill.Skill{{Name: name, Path: skillPath}}
				if err := adapter.Push(a.ctx, sk, agent.PushDir); err != nil {
					a.logErrorf("push starred skills to agents failed: agent=%s skill=%s err=%v", agentName, name, err)
					return nil, err
				}
			}
		}
	}
	a.logInfof("push starred skills to agents completed: conflicts=%d", len(conflicts))
	return conflicts, nil
}

// PushStarSkillsToAgentsForce copies starred skill directories to agent push directories,
// overwriting any existing skills.
func (a *App) PushStarSkillsToAgentsForce(skillPaths []string, agentNames []string) error {
	a.logInfof("force push starred skills to agents started: skillCount=%d agentCount=%d", len(skillPaths), len(agentNames))
	cfg, _ := a.config.Load()
	for _, agentName := range agentNames {
		for _, agent := range cfg.Agents {
			if agent.Name != agentName {
				continue
			}
			if agent.PushDir == "" {
				err := fmt.Errorf("智能体 %s 未配置推送路径", agentName)
				a.logErrorf("force push starred skills to agents failed: agent=%s err=%v", agentName, err)
				return err
			}
			var tempSkills []*skill.Skill
			for _, skillPath := range skillPaths {
				name := filepath.Base(skillPath)
				targetPath := filepath.Join(agent.PushDir, name)
				if err := os.RemoveAll(targetPath); err != nil {
					a.logErrorf("force push starred skills to agents failed: agent=%s target=%s err=%v", agentName, targetPath, err)
					return err
				}
				tempSkills = append(tempSkills, &skill.Skill{Name: name, Path: skillPath})
			}
			if err := getAdapter(agent).Push(a.ctx, tempSkills, agent.PushDir); err != nil {
				a.logErrorf("force push starred skills to agents failed: agent=%s err=%v", agentName, err)
				return err
			}
		}
	}
	a.logInfof("force push starred skills to agents completed")
	return nil
}

// --- Install ---

// ScanGitHub scans a remote git repo for valid skills, marking already-installed ones.
func (a *App) ScanGitHub(repoURL string) ([]install.SkillCandidate, error) {
	dataDir := config.AppDataDir()
	cacheDir, err := coregit.CacheDir(dataDir, repoURL)
	if err != nil {
		return nil, err
	}
	if err := coregit.CloneOrUpdate(a.ctx, repoURL, cacheDir, a.gitProxyURL()); err != nil {
		return nil, err
	}
	repoName, _ := coregit.ParseRepoName(repoURL)
	repoSource, err := coregit.RepoSource(repoURL)
	if err != nil {
		return nil, err
	}
	starSkills, err := coregit.ScanSkillsWithMaxDepth(cacheDir, repoURL, repoName, repoSource, a.repoScanMaxDepth())
	if err != nil {
		return nil, err
	}
	_, installedIndex, err := a.installedIndex()
	if err != nil {
		return nil, err
	}
	presence := a.buildAgentPresenceIndex(installedIndex)
	var candidates []install.SkillCandidate
	for _, ss := range starSkills {
		candidates = append(candidates, install.SkillCandidate{
			Name:       ss.Name,
			Path:       ss.SubPath,
			LogicalKey: skillkey.Git(repoSource, ss.SubPath),
		})
	}
	return resolveGitHubCandidates(candidates, installedIndex, presence), nil
}

// InstallFromGitHub imports selected skills from a scanned remote git repo into storage.
func (a *App) InstallFromGitHub(repoURL string, candidates []install.SkillCandidate, category string) error {
	category = normalizeCategoryName(category)
	dataDir := config.AppDataDir()
	cacheDir, err := coregit.CacheDir(dataDir, repoURL)
	if err != nil {
		return err
	}
	canonicalRepoURL, err := coregit.CanonicalRepoURL(repoURL)
	if err != nil {
		return err
	}
	_, installedIndex, err := a.installedIndex()
	if err != nil {
		return err
	}
	imported := make([]*skill.Skill, 0, len(candidates))
	for _, c := range candidates {
		logicalKey := c.LogicalKey
		if logicalKey == "" {
			logicalKey = skillkey.GitFromRepoURLOrEmpty(canonicalRepoURL, c.Path)
		}
		if installedIndex.IsInstalled(c.Name, logicalKey) {
			continue
		}
		skillDir := filepath.Join(cacheDir, filepath.FromSlash(c.Path))
		sha, _ := coregit.GetSubPathSHA(a.ctx, cacheDir, c.Path)
		sk, err := a.storage.Import(skillDir, category, skill.SourceGitHub, canonicalRepoURL, c.Path)
		if err != nil {
			return fmt.Errorf("import %s: %w", c.Name, err)
		}
		sk.SourceSHA = sha
		_ = a.storage.UpdateMeta(sk)
		imported = append(imported, sk)
	}
	a.autoPushImportedSkillsToAgents("github.install", imported)
	a.scheduleAutoBackup()
	return nil
}

func (a *App) ImportLocal(dir, category string) (*skill.Skill, error) {
	category = normalizeCategoryName(category)
	sk, err := a.storage.Import(dir, category, skill.SourceManual, "", "")
	if err != nil {
		return nil, err
	}
	a.autoPushImportedSkillsToAgents("local.import", []*skill.Skill{sk})
	a.scheduleAutoBackup()
	return sk, nil
}

func (a *App) autoPushAgentTargets(cfg config.AppConfig) []config.AgentConfig {
	if len(cfg.AutoPushAgents) == 0 {
		return nil
	}
	selected := make(map[string]struct{}, len(cfg.AutoPushAgents))
	for _, name := range cfg.AutoPushAgents {
		selected[name] = struct{}{}
	}
	targets := make([]config.AgentConfig, 0, len(selected))
	for _, agent := range cfg.Agents {
		if !agent.Enabled {
			continue
		}
		if _, ok := selected[agent.Name]; ok {
			targets = append(targets, agent)
		}
	}
	return targets
}

func (a *App) autoPushImportedSkillsToAgents(source string, imported []*skill.Skill) {
	a.autoPushSkillsToConfiguredAgents(source, imported, false)
}

func (a *App) autoPushSkillsToConfiguredAgents(source string, skills []*skill.Skill, overwriteExisting bool) {
	if len(skills) == 0 {
		return
	}
	cfg, err := a.config.Load()
	if err != nil {
		a.logErrorf("auto push imported skills failed: source=%s load config failed: %v", source, err)
		return
	}
	targets := a.autoPushAgentTargets(cfg)
	if len(targets) == 0 {
		a.logDebugf("auto push imported skills skipped: source=%s reason=no-target-agents", source)
		return
	}

	a.logInfof("auto push imported skills started: source=%s skillCount=%d agentCount=%d overwriteExisting=%t", source, len(skills), len(targets), overwriteExisting)
	totalPushed := 0
	failures := 0

	for _, agent := range targets {
		if strings.TrimSpace(agent.PushDir) == "" {
			failures++
			err := fmt.Errorf("push dir not configured")
			a.logErrorf("auto push agent failed: source=%s agent=%s err=%v", source, agent.Name, err)
			continue
		}

		toPush := make([]*skill.Skill, 0, len(skills))
		skippedExisting := 0
		for _, sk := range skills {
			targetPath := filepath.Join(agent.PushDir, sk.Name)
			if _, statErr := os.Stat(targetPath); statErr == nil {
				if overwriteExisting {
					if err := os.RemoveAll(targetPath); err != nil {
						failures++
						a.logErrorf("auto push agent failed: source=%s agent=%s skill=%s target=%s err=%v", source, agent.Name, sk.Name, targetPath, err)
						continue
					}
					toPush = append(toPush, sk)
					continue
				}
				skippedExisting++
				a.logDebugf("auto push agent skipped existing target: source=%s agent=%s skill=%s target=%s", source, agent.Name, sk.Name, targetPath)
				continue
			} else if !os.IsNotExist(statErr) {
				skippedExisting++
				failures++
				a.logErrorf("auto push agent failed: source=%s agent=%s skill=%s target=%s err=%v", source, agent.Name, sk.Name, targetPath, statErr)
				continue
			}
			toPush = append(toPush, sk)
		}

		a.logInfof("auto push agent started: source=%s agent=%s skillCount=%d", source, agent.Name, len(toPush))
		if len(toPush) == 0 {
			a.logInfof("auto push agent completed: source=%s agent=%s pushed=%d skippedExisting=%d", source, agent.Name, 0, skippedExisting)
			continue
		}
		if err := getAdapter(agent).Push(a.ctx, toPush, agent.PushDir); err != nil {
			failures++
			a.logErrorf("auto push agent failed: source=%s agent=%s pushDir=%s err=%v", source, agent.Name, agent.PushDir, err)
			continue
		}
		totalPushed += len(toPush)
		a.logInfof("auto push agent completed: source=%s agent=%s pushed=%d skippedExisting=%d", source, agent.Name, len(toPush), skippedExisting)
	}

	a.logInfof("auto push imported skills completed: source=%s skillCount=%d agentCount=%d pushed=%d failures=%d overwriteExisting=%t", source, len(skills), len(targets), totalPushed, failures, overwriteExisting)
}

// --- Sync ---

func (a *App) GetEnabledAgents() ([]config.AgentConfig, error) {
	cfg, err := a.config.Load()
	if err != nil {
		return nil, err
	}
	var enabled []config.AgentConfig
	for _, t := range cfg.Agents {
		if t.Enabled {
			enabled = append(enabled, t)
		}
	}
	return enabled, nil
}

// ScanAgentSkills lists all skills in an agent's configured scan directories for the pull page.
func (a *App) ScanAgentSkills(agentName string) ([]AgentSkillCandidate, error) {
	cfg, _ := a.config.Load()
	for _, agent := range cfg.Agents {
		if agent.Name == agentName {
			scanned, err := scanAgentSkillsRaw(a.ctx, getAdapter(agent), agent.ScanDirs, a.repoScanMaxDepth())
			if err != nil {
				return nil, err
			}
			_, installedIndex, err := a.installedIndex()
			if err != nil {
				return nil, err
			}
			return resolveAgentSkillCandidates(scanned, installedIndex, a.buildAgentPresenceIndex(installedIndex)), nil
		}
	}
	return nil, nil
}

// ListAgentSkills returns all skills for an agent, annotated with whether each
// skill lives in the push directory and/or the scan directories.
func (a *App) ListAgentSkills(agentName string) ([]AgentSkillEntry, error) {
	cfg, err := a.config.Load()
	if err != nil {
		return nil, err
	}
	var agentCfg *config.AgentConfig
	for i := range cfg.Agents {
		if cfg.Agents[i].Name == agentName {
			agentCfg = &cfg.Agents[i]
			break
		}
	}
	if agentCfg == nil {
		return nil, fmt.Errorf("agent %s not found", agentName)
	}
	adapter := getAdapter(*agentCfg)
	_, installedIndex, err := a.installedIndex()
	if err != nil {
		return nil, err
	}
	presence := a.buildAgentPresenceIndex(installedIndex)
	var pushSkills []*skill.Skill

	if agentCfg.PushDir != "" {
		if _, statErr := os.Stat(agentCfg.PushDir); statErr == nil {
			if pulled, pullErr := pullAgentSkills(a.ctx, adapter, agentCfg.PushDir, a.repoScanMaxDepth()); pullErr == nil {
				pushSkills = pulled
			}
		}
	}

	scanSkills, err := scanAgentSkillsRaw(a.ctx, adapter, agentCfg.ScanDirs, a.repoScanMaxDepth())
	if err != nil {
		return nil, err
	}
	return aggregateAgentSkillEntries(pushSkills, scanSkills, installedIndex, presence), nil
}

// DeleteAgentSkill removes a skill directory from an agent's push directory.
// Returns an error if skillPath is not within the agent's configured push directory.
func (a *App) DeleteAgentSkill(agentName string, skillPath string) error {
	cfg, err := a.config.Load()
	if err != nil {
		return err
	}
	var agentCfg *config.AgentConfig
	for i := range cfg.Agents {
		if cfg.Agents[i].Name == agentName {
			agentCfg = &cfg.Agents[i]
			break
		}
	}
	if agentCfg == nil {
		return fmt.Errorf("agent %s not found", agentName)
	}
	if agentCfg.PushDir == "" {
		return fmt.Errorf("智能体 %s 未配置推送路径", agentName)
	}
	rel, relErr := filepath.Rel(agentCfg.PushDir, skillPath)
	if relErr != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("无法删除不在推送路径下的 Skill")
	}
	a.logInfof("DeleteAgentSkill: deleting %s from agent %s push dir started", filepath.Base(skillPath), agentName)
	if err := os.RemoveAll(skillPath); err != nil {
		a.logErrorf("DeleteAgentSkill: delete %s failed: %v", skillPath, err)
		return fmt.Errorf("删除失败: %w", err)
	}
	a.logInfof("DeleteAgentSkill: deleted %s from agent %s push dir completed", filepath.Base(skillPath), agentName)
	return nil
}

// CheckMissingAgentPushDirs returns agent names and paths whose push directory does not yet exist.
// Each element is map{"name": agentName, "dir": pushDir}.
func (a *App) CheckMissingAgentPushDirs(agentNames []string) ([]map[string]string, error) {
	cfg, _ := a.config.Load()
	var missing []map[string]string
	for _, agentName := range agentNames {
		for _, agent := range cfg.Agents {
			if agent.Name != agentName || agent.PushDir == "" {
				continue
			}
			if _, err := os.Stat(agent.PushDir); os.IsNotExist(err) {
				missing = append(missing, map[string]string{"name": agent.Name, "dir": agent.PushDir})
			}
		}
	}
	return missing, nil
}

// PushToAgents pushes selected skills to target agents.
// Returns structured conflicts that were skipped.
func (a *App) PushToAgents(skillIDs []string, agentNames []string) ([]PushConflict, error) {
	a.logInfof("push skills to agents started: skillCount=%d agentCount=%d", len(skillIDs), len(agentNames))
	cfg, _ := a.config.Load()
	skills, err := a.storage.ListAll()
	if err != nil {
		a.logErrorf("push skills to agents failed: %v", err)
		return nil, err
	}
	idSet := map[string]bool{}
	for _, id := range skillIDs {
		idSet[id] = true
	}
	var selected []*skill.Skill
	for _, sk := range skills {
		if idSet[sk.ID] {
			selected = append(selected, sk)
		}
	}

	var conflicts []PushConflict
	for _, agentName := range agentNames {
		for _, agent := range cfg.Agents {
			if agent.Name != agentName {
				continue
			}
			if agent.PushDir == "" {
				err := fmt.Errorf("智能体 %s 未配置推送路径", agentName)
				a.logErrorf("push skills to agents failed: agent=%s err=%v", agentName, err)
				return nil, err
			}
			adapter := getAdapter(agent)
			var toPush []*skill.Skill
			for _, sk := range selected {
				dst := filepath.Join(agent.PushDir, sk.Name)
				if _, err := os.Stat(dst); err == nil {
					conflicts = append(conflicts, PushConflict{
						SkillID:    sk.ID,
						SkillName:  sk.Name,
						SkillPath:  sk.Path,
						AgentName:  agentName,
						TargetPath: dst,
					})
					continue
				}
				toPush = append(toPush, sk)
			}
			if len(toPush) == 0 {
				continue
			}
			if err := adapter.Push(a.ctx, toPush, agent.PushDir); err != nil {
				a.logErrorf("push skills to agents failed: agent=%s err=%v", agentName, err)
				return nil, err
			}
		}
	}
	a.logInfof("push skills to agents completed: conflicts=%d", len(conflicts))
	return conflicts, nil
}

// PushToAgentsForce pushes and overwrites conflicts.
func (a *App) PushToAgentsForce(skillIDs []string, agentNames []string) error {
	a.logInfof("force push skills to agents started: skillCount=%d agentCount=%d", len(skillIDs), len(agentNames))
	cfg, _ := a.config.Load()
	skills, _ := a.storage.ListAll()
	idSet := map[string]bool{}
	for _, id := range skillIDs {
		idSet[id] = true
	}
	var selected []*skill.Skill
	for _, sk := range skills {
		if idSet[sk.ID] {
			selected = append(selected, sk)
		}
	}
	for _, agentName := range agentNames {
		for _, agent := range cfg.Agents {
			if agent.Name == agentName {
				if agent.PushDir == "" {
					err := fmt.Errorf("智能体 %s 未配置推送路径", agentName)
					a.logErrorf("force push skills to agents failed: agent=%s err=%v", agentName, err)
					return err
				}
				for _, sk := range selected {
					targetPath := filepath.Join(agent.PushDir, sk.Name)
					if err := os.RemoveAll(targetPath); err != nil {
						a.logErrorf("force push skills to agents failed: agent=%s target=%s err=%v", agentName, targetPath, err)
						return err
					}
				}
				if err := getAdapter(agent).Push(a.ctx, selected, agent.PushDir); err != nil {
					a.logErrorf("force push skills to agents failed: agent=%s err=%v", agentName, err)
					return err
				}
			}
		}
	}
	a.logInfof("force push skills to agents completed")
	return nil
}

// PullFromAgent imports selected skills from an agent into SkillFlow storage.
func (a *App) PullFromAgent(agentName string, skillPaths []string, category string) ([]string, error) {
	category = normalizeCategoryName(category)
	cfg, _ := a.config.Load()
	for _, agent := range cfg.Agents {
		if agent.Name != agentName {
			continue
		}
		scanned, err := scanAgentSkillsRaw(a.ctx, getAdapter(agent), agent.ScanDirs, a.repoScanMaxDepth())
		if err != nil {
			return nil, err
		}
		_, installedIndex, err := a.installedIndex()
		if err != nil {
			return nil, err
		}
		candidates := resolveAgentSkillSelection(resolveAgentSkillCandidates(scanned, installedIndex, a.buildAgentPresenceIndex(installedIndex)), skillPaths)
		var conflicts []string
		imported := make([]*skill.Skill, 0, len(candidates))
		for _, candidate := range candidates {
			if candidate.Imported {
				conflicts = append(conflicts, candidate.Path)
				continue
			}
			sk, err := a.storage.Import(candidate.Path, category, skill.SourceManual, "", "")
			if err == skill.ErrSkillExists {
				conflicts = append(conflicts, candidate.Path)
				continue
			}
			if err != nil {
				return nil, err
			}
			imported = append(imported, sk)
		}
		a.autoPushImportedSkillsToAgents("agent.pull", imported)
		a.scheduleAutoBackup()
		return conflicts, nil
	}
	return nil, nil
}

// PullFromAgentForce imports selected skills, overwriting existing ones.
func (a *App) PullFromAgentForce(agentName string, skillPaths []string, category string) error {
	category = normalizeCategoryName(category)
	cfg, _ := a.config.Load()
	for _, agent := range cfg.Agents {
		if agent.Name != agentName {
			continue
		}
		scanned, err := scanAgentSkillsRaw(a.ctx, getAdapter(agent), agent.ScanDirs, a.repoScanMaxDepth())
		if err != nil {
			return err
		}
		_, installedIndex, err := a.installedIndex()
		if err != nil {
			return err
		}
		imported := make([]*skill.Skill, 0, len(skillPaths))
		for _, candidate := range resolveAgentSkillSelection(resolveAgentSkillCandidates(scanned, installedIndex, a.buildAgentPresenceIndex(installedIndex)), skillPaths) {
			existing, _ := a.storage.ListAll()
			for _, e := range existing {
				logicalKey, logicalErr := skill.LogicalKey(e)
				if (candidate.LogicalKey != "" && logicalErr == nil && logicalKey == candidate.LogicalKey) || (candidate.LogicalKey == "" && e.Name == candidate.Name) {
					_ = a.storage.Delete(e.ID)
				}
			}
			sk, err := a.storage.Import(candidate.Path, category, skill.SourceManual, "", "")
			if err != nil {
				return err
			}
			imported = append(imported, sk)
		}
		a.autoPushImportedSkillsToAgents("agent.pull.force", imported)
		a.scheduleAutoBackup()
	}
	return nil
}

func pullAgentSkills(ctx context.Context, adapter agentsync.AgentAdapter, dir string, maxDepth int) ([]*skill.Skill, error) {
	if depthAware, ok := adapter.(maxDepthPuller); ok {
		return depthAware.PullWithMaxDepth(ctx, dir, maxDepth)
	}
	return adapter.Pull(ctx, dir)
}

func getAdapter(t config.AgentConfig) agentsync.AgentAdapter {
	if a, ok := registry.GetAdapter(t.Name); ok {
		return a
	}
	return agentsync.NewFilesystemAdapter(t.Name, t.PushDir)
}

func (a *App) repoScanMaxDepth() int {
	cfg, err := a.config.Load()
	if err != nil {
		return config.DefaultRepoScanMaxDepth
	}
	return config.NormalizeRepoScanMaxDepth(cfg.RepoScanMaxDepth)
}

// --- Config ---

func (a *App) GetConfig() (config.AppConfig, error) {
	cfg, err := a.config.Load()
	if err != nil {
		return cfg, err
	}
	cfg.DefaultCategory = defaultCategoryName
	cfg.LogLevel = config.NormalizeLogLevel(cfg.LogLevel)
	cfg.RepoScanMaxDepth = config.NormalizeRepoScanMaxDepth(cfg.RepoScanMaxDepth)
	return cfg, nil
}

func (a *App) SaveConfig(cfg config.AppConfig) error {
	a.logInfof("save config requested")
	prevCfg, err := a.config.Load()
	if err != nil {
		a.logErrorf("save config failed: load current config failed: %v", err)
		return err
	}
	cfg.DefaultCategory = defaultCategoryName
	cfg.LogLevel = config.NormalizeLogLevel(cfg.LogLevel)
	cfg.RepoScanMaxDepth = config.NormalizeRepoScanMaxDepth(cfg.RepoScanMaxDepth)
	if err := a.config.Save(cfg); err != nil {
		a.logErrorf("save config failed: %v", err)
		return err
	}
	a.syncExistingSkillsForNewAutoPushAgents(prevCfg, cfg)
	if prevCfg.LaunchAtLogin != cfg.LaunchAtLogin {
		if err := a.syncLaunchAtLogin(cfg.LaunchAtLogin); err != nil {
			persistedCfg := cfg
			rollbackCfg := cfg
			rollbackCfg.LaunchAtLogin = prevCfg.LaunchAtLogin
			if rollbackErr := a.config.Save(rollbackCfg); rollbackErr != nil {
				a.logErrorf("save config rollback failed: restore launch-at-login setting=%t err=%v", prevCfg.LaunchAtLogin, rollbackErr)
			} else if rollbackErr := a.syncLaunchAtLogin(prevCfg.LaunchAtLogin); rollbackErr != nil {
				a.logErrorf("save config rollback failed: restore launch-at-login state=%t err=%v", prevCfg.LaunchAtLogin, rollbackErr)
				persistedCfg = rollbackCfg
			} else {
				persistedCfg = rollbackCfg
			}
			a.setLoggerLevel(persistedCfg.LogLevel)
			a.startAutoSyncTimer(persistedCfg.Cloud.SyncIntervalMinutes)
			a.logErrorf("save config failed: apply launch-at-login failed: %v", err)
			return err
		}
	}
	a.setLoggerLevel(cfg.LogLevel)
	a.logInfof("save config completed: logLevel=%s repoScanMaxDepth=%d launchAtLogin=%t", cfg.LogLevel, cfg.RepoScanMaxDepth, cfg.LaunchAtLogin)
	a.startAutoSyncTimer(cfg.Cloud.SyncIntervalMinutes)
	return nil
}

func (a *App) syncExistingSkillsForNewAutoPushAgents(prevCfg, nextCfg config.AppConfig) {
	prevTargets := make(map[string]struct{}, len(prevCfg.AutoPushAgents))
	for _, name := range prevCfg.AutoPushAgents {
		prevTargets[name] = struct{}{}
	}

	newTargets := make([]string, 0, len(nextCfg.AutoPushAgents))
	for _, name := range nextCfg.AutoPushAgents {
		if _, existed := prevTargets[name]; existed {
			continue
		}
		newTargets = append(newTargets, name)
	}
	if len(newTargets) == 0 {
		return
	}

	skills, err := a.storage.ListAll()
	if err != nil {
		a.logErrorf("sync existing skills to new auto push agents failed: load skills failed: %v", err)
		return
	}
	if len(skills) == 0 {
		a.logInfof("sync existing skills to new auto push agents skipped: reason=no-skills agentCount=%d", len(newTargets))
		return
	}

	skillIDs := make([]string, 0, len(skills))
	for _, sk := range skills {
		skillIDs = append(skillIDs, sk.ID)
	}

	a.logInfof("sync existing skills to new auto push agents started: skillCount=%d agentCount=%d", len(skillIDs), len(newTargets))
	conflicts, err := a.PushToAgents(skillIDs, newTargets)
	if err != nil {
		a.logErrorf("sync existing skills to new auto push agents failed: %v", err)
		return
	}
	a.logInfof("sync existing skills to new auto push agents completed: skillCount=%d agentCount=%d conflicts=%d", len(skillIDs), len(newTargets), len(conflicts))
}

func (a *App) AddCustomAgent(name, pushDir string) error {
	a.logInfof("add custom agent requested: name=%s", name)
	cfg, err := a.config.Load()
	if err != nil {
		a.logErrorf("add custom agent failed: %v", err)
		return err
	}
	cfg.Agents = append(cfg.Agents, config.AgentConfig{
		Name:     name,
		ScanDirs: []string{pushDir},
		PushDir:  pushDir,
		Enabled:  true,
		Custom:   true,
	})
	if err := a.config.Save(cfg); err != nil {
		a.logErrorf("add custom agent failed: %v", err)
		return err
	}
	a.logInfof("add custom agent done: name=%s", name)
	return nil
}

func (a *App) RemoveCustomAgent(name string) error {
	a.logInfof("remove custom agent requested: name=%s", name)
	cfg, err := a.config.Load()
	if err != nil {
		a.logErrorf("remove custom agent failed: %v", err)
		return err
	}
	var filtered []config.AgentConfig
	for _, t := range cfg.Agents {
		if !(t.Custom && t.Name == name) {
			filtered = append(filtered, t)
		}
	}
	cfg.Agents = filtered
	if err := a.config.Save(cfg); err != nil {
		a.logErrorf("remove custom agent failed: %v", err)
		return err
	}
	a.logInfof("remove custom agent done: name=%s", name)
	return nil
}

// --- Backup ---

func (a *App) BackupNow() error {
	a.logInfof("manual backup requested")
	return a.runBackup()
}

func (a *App) ListCloudFiles() ([]backup.RemoteFile, error) {
	cfg, err := a.config.Load()
	if err != nil {
		a.logErrorf("list cloud files failed: load config failed: %v", err)
		return nil, err
	}
	a.logInfof("list cloud files started (provider=%s, remotePath=%s)", cfg.Cloud.Provider, cfg.Cloud.RemotePath)
	provider, ok := registry.GetCloudProvider(cfg.Cloud.Provider)
	if !ok {
		err := fmt.Errorf("provider not found: %s", cfg.Cloud.Provider)
		a.logErrorf("list cloud files failed: %v", err)
		return nil, err
	}
	if err := provider.Init(cfg.Cloud.Credentials); err != nil {
		a.logErrorf("list cloud files failed: init provider %s failed: %v", cfg.Cloud.Provider, err)
		return nil, err
	}
	files, err := provider.List(a.ctx, cfg.Cloud.BucketName, cfg.Cloud.RemotePath)
	if err != nil {
		a.logErrorf("list cloud files failed: provider=%s, remotePath=%s, err=%v", cfg.Cloud.Provider, cfg.Cloud.RemotePath, err)
		return nil, err
	}
	a.logInfof("list cloud files completed (provider=%s, remotePath=%s, count=%d)", cfg.Cloud.Provider, cfg.Cloud.RemotePath, len(files))
	return files, nil
}

func (a *App) RestoreFromCloud() error {
	a.logInfof("restore from cloud requested")
	cfg, err := a.config.Load()
	if err != nil {
		a.logErrorf("restore from cloud failed: %v", err)
		return err
	}
	beforeRestore := a.captureCloudRestoreState()
	provider, ok := registry.GetCloudProvider(cfg.Cloud.Provider)
	if !ok {
		return fmt.Errorf("provider not found: %s", cfg.Cloud.Provider)
	}
	if err := provider.Init(cfg.Cloud.Credentials); err != nil {
		return err
	}
	isGit := cfg.Cloud.Provider == backup.GitProviderName
	restoreDir := a.backupRootDir(cfg)
	if isGit {
		restoreDir, err = a.prepareGitBackupRoot(cfg)
		if err != nil {
			a.logErrorf("restore from cloud failed: prepare git backup root failed: %v", err)
			return err
		}
	}
	beforeSnapshot, err := backup.BuildSnapshot(restoreDir)
	if err != nil {
		a.logErrorf("restore from cloud failed: build pre-restore snapshot failed: %v", err)
		return err
	}
	if isGit {
		a.hub.Publish(notify.Event{Type: notify.EventGitSyncStarted})
	}
	if err := provider.Restore(a.ctx, cfg.Cloud.BucketName, cfg.Cloud.RemotePath, restoreDir); err != nil {
		a.logErrorf("restore from cloud failed: %v", err)
		var conflictErr *backup.GitConflictError
		if isGit && errors.As(err, &conflictErr) {
			a.publishGitConflict(conflictErr)
		}
		if isGit {
			a.hub.Publish(notify.Event{Type: notify.EventGitSyncFailed, Payload: err.Error()})
		}
		return err
	}
	changes, afterSnapshot, err := a.computeSnapshotChanges(beforeSnapshot, restoreDir)
	if err != nil {
		a.logErrorf("restore from cloud failed: compute post-restore changes failed: %v", err)
		if isGit {
			a.hub.Publish(notify.Event{Type: notify.EventGitSyncFailed, Payload: err.Error()})
		}
		return err
	}
	a.recordBackupResult(changes, afterSnapshot)
	a.logInfof("restore from cloud completed")
	if err := a.handleRestoredCloudState(beforeRestore, "cloud.restore"); err != nil {
		a.logErrorf("restore from cloud failed: restore compensation failed: %v", err)
		if isGit {
			a.hub.Publish(notify.Event{Type: notify.EventGitSyncFailed, Payload: err.Error()})
		}
		return err
	}
	if isGit {
		a.clearGitConflictPending()
		a.hub.Publish(notify.Event{Type: notify.EventGitSyncCompleted, Payload: notify.BackupCompletedPayload{Files: changes, CompletedAt: a.GetLastBackupCompletedAt()}})
	}
	return nil
}

// ListCloudProviders returns all registered provider names and their required credential fields.
func (a *App) ListCloudProviders() []map[string]any {
	var result []map[string]any
	for _, p := range registry.AllCloudProviders() {
		result = append(result, map[string]any{
			"name":   p.Name(),
			"fields": p.RequiredCredentials(),
		})
	}
	return result
}

// --- Updates ---

func (a *App) CheckUpdates() error {
	a.logInfof("check skill updates started")
	skills, err := a.storage.ListAll()
	if err != nil {
		a.logErrorf("check skill updates failed: %v", err)
		return err
	}

	type updateGroup struct {
		LogicalKey string
		Skills     []*skill.Skill
	}

	groups := map[string]*updateGroup{}
	for _, sk := range skills {
		if sk == nil || !sk.IsGitHub() {
			continue
		}
		logicalKey, logicalErr := skillkey.GitFromRepoURL(sk.SourceURL, sk.SourceSubPath)
		if logicalErr != nil || strings.TrimSpace(logicalKey) == "" {
			if logicalErr != nil {
				a.logErrorf("check skill updates failed: skillID=%s name=%s err=%v", sk.ID, sk.Name, logicalErr)
			}
			continue
		}
		group := groups[logicalKey]
		if group == nil {
			group = &updateGroup{LogicalKey: logicalKey}
			groups[logicalKey] = group
		}
		group.Skills = append(group.Skills, sk)
	}

	for _, group := range groups {
		reference := group.Skills[0]
		a.logInfof("check skill updates started: logicalKey=%s instances=%d", group.LogicalKey, len(group.Skills))
		checkedAt := time.Now()
		_, latestSHA, err := a.cachedSkillSourceDir(reference)
		if err != nil {
			a.logErrorf("check skill updates failed: logicalKey=%s repo=%s subPath=%s err=%v", group.LogicalKey, reference.SourceURL, reference.SourceSubPath, err)
			for _, sk := range group.Skills {
				sk.LastCheckedAt = checkedAt
				sk.LatestSHA = ""
				_ = a.storage.SaveMeta(sk)
			}
			continue
		}
		for _, sk := range group.Skills {
			sk.LastCheckedAt = checkedAt
			if latestSHA != "" && latestSHA != sk.SourceSHA {
				sk.LatestSHA = latestSHA
			} else {
				sk.LatestSHA = ""
			}
			_ = a.storage.SaveMeta(sk)
			if sk.LatestSHA != "" {
				a.hub.Publish(notify.Event{
					Type: notify.EventUpdateAvailable,
					Payload: notify.UpdateAvailablePayload{
						SkillID:    sk.ID,
						SkillName:  sk.Name,
						CurrentSHA: sk.SourceSHA,
						LatestSHA:  latestSHA,
					},
				})
			}
		}
		a.logInfof("check skill updates completed: logicalKey=%s latestSHA=%s", group.LogicalKey, latestSHA)
	}
	a.logInfof("check skill updates completed")
	return nil
}

// UpdateSkill refreshes an installed GitHub skill from the local cached repo copy.
func (a *App) UpdateSkill(skillID string) error {
	a.logInfof("update skill requested: id=%s", skillID)
	sk, err := a.storage.Get(skillID)
	if err != nil {
		a.logErrorf("update skill failed: %v", err)
		return err
	}
	cacheSourceDir, latestSHA, err := a.cachedSkillSourceDir(sk)
	if err != nil {
		a.logErrorf("update skill cache prepare failed: id=%s name=%s repo=%s subPath=%s err=%v", skillID, sk.Name, sk.SourceURL, sk.SourceSubPath, err)
		return err
	}
	a.logInfof("update skill cache copy started: id=%s name=%s source=%s", skillID, sk.Name, cacheSourceDir)
	if err := a.storage.OverwriteFromDir(skillID, cacheSourceDir); err != nil {
		a.logErrorf("update skill overwrite failed: %v", err)
		return err
	}
	sk.SourceSHA = latestSHA
	sk.LatestSHA = ""
	_ = a.storage.UpdateMeta(sk)
	if err := a.refreshUpdatedPushedSkillCopies(sk); err != nil {
		a.logErrorf("update skill refresh pushed copies failed: id=%s name=%s err=%v", skillID, sk.Name, err)
		a.scheduleAutoBackup()
		return err
	}
	a.autoPushSkillsToConfiguredAgents("skill.update", []*skill.Skill{sk}, true)
	a.scheduleAutoBackup()
	a.logInfof("update skill completed: id=%s name=%s", skillID, sk.Name)
	return nil
}

func (a *App) cachedSkillSourceDir(sk *skill.Skill) (string, string, error) {
	if sk == nil {
		return "", "", fmt.Errorf("skill is required")
	}
	if a.config == nil {
		return "", "", fmt.Errorf("config service is not initialized")
	}

	cacheDir, err := coregit.CacheDir(a.config.DataDir(), sk.SourceURL)
	if err != nil {
		return "", "", err
	}
	if _, err := os.Stat(filepath.Join(cacheDir, ".git")); err != nil {
		if os.IsNotExist(err) {
			return "", "", fmt.Errorf("local cache missing for repo: %s", sk.SourceURL)
		}
		return "", "", err
	}

	subPath := strings.TrimSpace(sk.SourceSubPath)
	sourceDir := cacheDir
	shaPath := "."
	if subPath != "" {
		sourceDir = filepath.Join(cacheDir, filepath.FromSlash(subPath))
		shaPath = filepath.ToSlash(filepath.Clean(filepath.FromSlash(subPath)))
	}
	info, err := os.Stat(sourceDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", fmt.Errorf("cached skill path missing: %s", sk.SourceSubPath)
		}
		return "", "", err
	}
	if !info.IsDir() {
		return "", "", fmt.Errorf("cached skill path is not a directory: %s", sk.SourceSubPath)
	}

	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	latestSHA, err := coregit.GetSubPathSHA(ctx, cacheDir, shaPath)
	if err != nil {
		return "", "", err
	}
	return sourceDir, latestSHA, nil
}

func (a *App) newGitHubDownloader() githubSkillDownloader {
	if a.ghDownloader != nil {
		return a.ghDownloader(a.proxyHTTPClient())
	}
	return install.NewGitHubInstaller("", a.proxyHTTPClient())
}

func (a *App) refreshUpdatedPushedSkillCopies(sk *skill.Skill) error {
	if sk == nil {
		return nil
	}

	cfg, err := a.config.Load()
	if err != nil {
		a.logErrorf("refresh updated pushed skill copies failed: skill=%s load config failed: %v", sk.Name, err)
		return err
	}

	a.logInfof("refresh updated pushed skill copies started: skill=%s", sk.Name)
	updatedAgents := 0

	for _, agent := range cfg.Agents {
		if strings.TrimSpace(agent.PushDir) == "" {
			continue
		}

		targetPath := filepath.Join(agent.PushDir, sk.Name)
		if _, err := os.Stat(targetPath); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			a.logErrorf("refresh updated pushed skill copy failed: skill=%s agent=%s target=%s err=%v", sk.Name, agent.Name, targetPath, err)
			return err
		}

		a.logInfof("refresh updated pushed skill copy started: skill=%s agent=%s target=%s", sk.Name, agent.Name, targetPath)
		if err := os.RemoveAll(targetPath); err != nil {
			a.logErrorf("refresh updated pushed skill copy failed: skill=%s agent=%s target=%s err=%v", sk.Name, agent.Name, targetPath, err)
			return err
		}
		if err := getAdapter(agent).Push(a.ctx, []*skill.Skill{sk}, agent.PushDir); err != nil {
			a.logErrorf("refresh updated pushed skill copy failed: skill=%s agent=%s target=%s err=%v", sk.Name, agent.Name, targetPath, err)
			return err
		}
		updatedAgents++
		a.logInfof("refresh updated pushed skill copy completed: skill=%s agent=%s target=%s", sk.Name, agent.Name, targetPath)
	}

	a.logInfof("refresh updated pushed skill copies completed: skill=%s updatedAgents=%d", sk.Name, updatedAgents)
	return nil
}

func (a *App) checkUpdatesOnStartup() {
	_ = a.CheckUpdates()
}

func (a *App) updateStarredReposOnStartup() {
	_ = a.UpdateAllStarredRepos()
}

// OpenFolderDialog wraps Wails file dialog for frontend use.
func (a *App) OpenFolderDialog(defaultDir string) (string, error) {
	options := runtime.OpenDialogOptions{Title: "选择 Skill 目录"}
	if dir := nearestExistingDirectory(defaultDir); dir != "" {
		options.DefaultDirectory = dir
	}
	return runtime.OpenDirectoryDialog(a.ctx, options)
}

// OpenPath opens the given filesystem path in the OS default file manager.
func (a *App) OpenPath(path string) error {
	target, err := resolveOpenPathTarget(path)
	if err != nil {
		a.logErrorf("open path failed: path=%s, err=%v", path, err)
		return err
	}
	a.logInfof("open path started: requested=%s target=%s", path, target)
	var cmd *exec.Cmd
	switch goruntime.GOOS {
	case "darwin":
		cmd = exec.Command("open", "--", target)
	case "windows":
		cmd = exec.Command("explorer.exe", filepath.Clean(target))
	default:
		cmd = exec.Command("xdg-open", target)
	}
	if err := cmd.Start(); err != nil {
		a.logErrorf("open path failed: requested=%s target=%s err=%v", path, target, err)
		return err
	}
	a.logInfof("open path completed: requested=%s target=%s", path, target)
	return nil
}

// GetLogDir returns the app log directory path.
func (a *App) GetLogDir() string {
	return a.logDir()
}

// OpenLogDir opens the app log directory in the OS file manager.
func (a *App) OpenLogDir() error {
	return a.OpenPath(a.logDir())
}

// Greet is kept for Wails template compatibility.
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// --- Starred Repos ---

func (a *App) AddStarredRepo(repoURL string) (*coregit.StarredRepo, error) {
	a.logInfof("add starred repo requested: %s", repoURL)
	if err := coregit.CheckGitInstalled(); err != nil {
		a.logErrorf("add starred repo failed: %v", err)
		return nil, err
	}
	repos, err := a.starStorage.Load()
	if err != nil {
		a.logErrorf("add starred repo failed: %v", err)
		return nil, err
	}
	for i, r := range repos {
		if coregit.SameRepo(r.URL, repoURL) {
			if repos[i].Source == "" {
				if source, err := coregit.RepoSource(repos[i].URL); err == nil {
					repos[i].Source = source
					_ = a.starStorage.Save(repos)
				}
			}
			return &repos[i], nil // already starred
		}
	}
	name, err := coregit.ParseRepoName(repoURL)
	if err != nil {
		a.logErrorf("add starred repo failed: %v", err)
		return nil, err
	}
	dataDir := filepath.Dir(a.cacheDir)
	localDir, err := coregit.CacheDir(dataDir, repoURL)
	if err != nil {
		a.logErrorf("add starred repo failed: %v", err)
		return nil, err
	}
	source, err := coregit.RepoSource(repoURL)
	if err != nil {
		a.logErrorf("add starred repo failed: %v", err)
		return nil, err
	}
	repo := coregit.StarredRepo{URL: repoURL, Name: name, Source: source, LocalDir: localDir}
	if cloneErr := coregit.CloneOrUpdate(a.ctx, repoURL, localDir, a.gitProxyURL()); cloneErr != nil {
		// Return typed errors for auth failures so the frontend can show the right dialog.
		if coregit.IsSSHAuthError(cloneErr) {
			return nil, fmt.Errorf("AUTH_SSH:%s", cloneErr.Error())
		}
		if coregit.IsAuthError(cloneErr) {
			return nil, fmt.Errorf("AUTH_HTTP:%s", cloneErr.Error())
		}
		repo.SyncError = cloneErr.Error()
	} else {
		repo.LastSync = time.Now()
	}
	repos = append(repos, repo)
	if err := a.starStorage.Save(repos); err != nil {
		a.logErrorf("add starred repo failed: %v", err)
		return nil, err
	}
	a.logInfof("add starred repo completed: %s", repoURL)
	return &repos[len(repos)-1], nil
}

// AddStarredRepoWithCredentials clones a repo using the provided HTTP username/password,
// removing any previously failed entry for the same URL first.
func (a *App) AddStarredRepoWithCredentials(repoURL, username, password string) (*coregit.StarredRepo, error) {
	a.logInfof("add starred repo with credentials requested: %s", repoURL)
	if err := coregit.CheckGitInstalled(); err != nil {
		a.logErrorf("add starred repo with credentials failed: %v", err)
		return nil, err
	}
	repos, err := a.starStorage.Load()
	if err != nil {
		a.logErrorf("add starred repo with credentials failed: %v", err)
		return nil, err
	}
	// Remove any existing (possibly failed) entry for this URL.
	filtered := repos[:0]
	for _, r := range repos {
		if !coregit.SameRepo(r.URL, repoURL) {
			filtered = append(filtered, r)
		}
	}
	name, err := coregit.ParseRepoName(repoURL)
	if err != nil {
		a.logErrorf("add starred repo with credentials failed: %v", err)
		return nil, err
	}
	dataDir := filepath.Dir(a.cacheDir)
	localDir, err := coregit.CacheDir(dataDir, repoURL)
	if err != nil {
		a.logErrorf("add starred repo with credentials failed: %v", err)
		return nil, err
	}
	source, err := coregit.RepoSource(repoURL)
	if err != nil {
		a.logErrorf("add starred repo with credentials failed: %v", err)
		return nil, err
	}
	repo := coregit.StarredRepo{URL: repoURL, Name: name, Source: source, LocalDir: localDir}
	if cloneErr := coregit.CloneOrUpdateWithCreds(a.ctx, repoURL, localDir, a.gitProxyURL(), username, password); cloneErr != nil {
		a.logErrorf("add starred repo with credentials failed: %v", cloneErr)
		return nil, cloneErr
	}
	repo.LastSync = time.Now()
	filtered = append(filtered, repo)
	if err := a.starStorage.Save(filtered); err != nil {
		a.logErrorf("add starred repo with credentials failed: %v", err)
		return nil, err
	}
	a.logInfof("add starred repo with credentials completed: %s", repoURL)
	return &filtered[len(filtered)-1], nil
}

func (a *App) RemoveStarredRepo(repoURL string) error {
	a.logInfof("remove starred repo requested: %s", repoURL)
	repos, err := a.starStorage.Load()
	if err != nil {
		a.logErrorf("remove starred repo failed: %v", err)
		return err
	}
	filtered := make([]coregit.StarredRepo, 0, len(repos))
	for _, r := range repos {
		if !coregit.SameRepo(r.URL, repoURL) {
			filtered = append(filtered, r)
		}
	}
	if err := a.starStorage.Save(filtered); err != nil {
		a.logErrorf("remove starred repo failed: %v", err)
		return err
	}
	a.logInfof("remove starred repo completed: %s", repoURL)
	return nil
}

func (a *App) ListStarredRepos() ([]coregit.StarredRepo, error) {
	repos, err := a.starStorage.Load()
	if repos == nil {
		return []coregit.StarredRepo{}, err
	}
	changed := false
	for i := range repos {
		if repos[i].Source != "" {
			continue
		}
		if source, parseErr := coregit.RepoSource(repos[i].URL); parseErr == nil {
			repos[i].Source = source
			changed = true
		}
	}
	if changed {
		_ = a.starStorage.Save(repos)
	}
	return repos, err
}

func (a *App) ListAllStarSkills() ([]coregit.StarSkill, error) {
	return measureOperation(a, "list_all_star_skills", func() ([]coregit.StarSkill, error) {
		fingerprint, fingerprintErr := a.allStarSkillsFingerprint()
		if fingerprintErr == nil {
			var cached []coregit.StarSkill
			state, err := a.ensureViewCache().Load(allStarSkillsSnapshotName, fingerprint, &cached)
			if err == nil && state == viewstate.StateHit {
				return cached, nil
			}
		}

		skills, err := a.listAllStarSkillsUncached()
		if err != nil {
			return nil, err
		}
		if fingerprint, err := a.allStarSkillsFingerprint(); err == nil {
			_ = a.ensureViewCache().Save(allStarSkillsSnapshotName, fingerprint, skills)
		}
		return skills, nil
	})
}

func (a *App) listAllStarSkillsUncached() ([]coregit.StarSkill, error) {
	repos, err := a.starStorage.Load()
	if err != nil {
		return nil, err
	}
	maxDepth := a.repoScanMaxDepth()
	_, installedIndex, err := a.installedIndex()
	if err != nil {
		return nil, err
	}
	presence := a.buildAgentPresenceIndex(installedIndex)
	var all []coregit.StarSkill
	for _, r := range repos {
		source := r.Source
		if source == "" {
			source, _ = coregit.RepoSource(r.URL)
		}
		skills, _ := coregit.ScanSkillsWithMaxDepth(r.LocalDir, r.URL, r.Name, source, maxDepth)
		all = append(all, resolveStarSkills(skills, installedIndex, presence)...)
	}
	if all == nil {
		return []coregit.StarSkill{}, nil
	}
	return all, nil
}

func (a *App) ListRepoStarSkills(repoURL string) ([]coregit.StarSkill, error) {
	repos, err := a.starStorage.Load()
	if err != nil {
		return nil, err
	}
	maxDepth := a.repoScanMaxDepth()
	_, installedIndex, err := a.installedIndex()
	if err != nil {
		return nil, err
	}
	presence := a.buildAgentPresenceIndex(installedIndex)
	for _, r := range repos {
		if !coregit.SameRepo(r.URL, repoURL) {
			continue
		}
		source := r.Source
		if source == "" {
			source, _ = coregit.RepoSource(r.URL)
		}
		skills, err := coregit.ScanSkillsWithMaxDepth(r.LocalDir, r.URL, r.Name, source, maxDepth)
		if err != nil {
			return nil, err
		}
		skills = resolveStarSkills(skills, installedIndex, presence)
		if skills == nil {
			return []coregit.StarSkill{}, nil
		}
		return skills, nil
	}
	return []coregit.StarSkill{}, nil
}

func (a *App) UpdateStarredRepo(repoURL string) error {
	a.logInfof("update starred repo started: repo=%s", repoURL)
	repos, err := a.starStorage.Load()
	if err != nil {
		a.logErrorf("update starred repo failed: repo=%s err=%v", repoURL, err)
		return err
	}
	for i, r := range repos {
		if !coregit.SameRepo(r.URL, repoURL) {
			continue
		}
		syncErr := coregit.CloneOrUpdate(a.ctx, r.URL, r.LocalDir, a.gitProxyURL())
		if syncErr != nil {
			a.logErrorf("update starred repo failed: repo=%s err=%v", r.URL, syncErr)
			repos[i].SyncError = syncErr.Error()
		} else {
			a.logInfof("update starred repo completed: repo=%s", r.URL)
			repos[i].SyncError = ""
			repos[i].LastSync = time.Now()
		}
		err = a.starStorage.Save(repos)
		if err != nil {
			a.logErrorf("update starred repo failed: repo=%s err=%v", r.URL, err)
			return err
		}
		a.hub.Publish(notify.Event{
			Type: notify.EventStarSyncProgress,
			Payload: notify.StarSyncProgressPayload{
				RepoURL:   repos[i].URL,
				RepoName:  repos[i].Name,
				SyncError: repos[i].SyncError,
			},
		})
		return nil
	}
	a.logErrorf("update starred repo failed: repo=%s err=starred repo not found", repoURL)
	return nil
}

func (a *App) UpdateAllStarredRepos() error {
	a.logInfof("update all starred repos requested")
	repos, err := a.starStorage.Load()
	if err != nil {
		a.logErrorf("update all starred repos failed: %v", err)
		return err
	}
	if len(repos) == 0 {
		return nil
	}
	var wg sync.WaitGroup
	mu := &sync.Mutex{}
	for i := range repos {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			r := repos[idx]
			a.logInfof("update starred repo started: repo=%s", r.URL)
			syncErr := coregit.CloneOrUpdate(a.ctx, r.URL, r.LocalDir, a.gitProxyURL())
			mu.Lock()
			if syncErr != nil {
				a.logErrorf("update starred repo failed: repo=%s err=%v", r.URL, syncErr)
				repos[idx].SyncError = syncErr.Error()
			} else {
				a.logInfof("update starred repo completed: repo=%s", r.URL)
				repos[idx].SyncError = ""
				repos[idx].LastSync = time.Now()
			}
			mu.Unlock()
			a.hub.Publish(notify.Event{
				Type: notify.EventStarSyncProgress,
				Payload: notify.StarSyncProgressPayload{
					RepoURL:   repos[idx].URL,
					RepoName:  repos[idx].Name,
					SyncError: repos[idx].SyncError,
				},
			})
		}(i)
	}
	wg.Wait()
	a.hub.Publish(notify.Event{Type: notify.EventStarSyncDone})
	if err := a.starStorage.Save(repos); err != nil {
		a.logErrorf("update all starred repos failed: %v", err)
		return err
	}
	a.logInfof("update all starred repos completed")
	return nil
}

func (a *App) ImportStarSkills(skillPaths []string, repoURL, category string) error {
	category = normalizeCategoryName(category)
	repos, _ := a.starStorage.Load()
	var repoLocalDir string
	canonicalRepoURL := repoURL
	if normalized, err := coregit.CanonicalRepoURL(repoURL); err == nil {
		canonicalRepoURL = normalized
	}
	for _, r := range repos {
		if coregit.SameRepo(r.URL, repoURL) {
			repoLocalDir = r.LocalDir
			if normalized, err := coregit.CanonicalRepoURL(r.URL); err == nil {
				canonicalRepoURL = normalized
			}
			break
		}
	}
	if repoLocalDir == "" {
		return fmt.Errorf("starred repo not found: %s", repoURL)
	}
	_, installedIndex, err := a.installedIndex()
	if err != nil {
		return err
	}
	imported := make([]*skill.Skill, 0, len(skillPaths))
	for _, skillPath := range skillPaths {
		subPath, _ := filepath.Rel(repoLocalDir, skillPath)
		subPath = filepath.ToSlash(subPath)
		logicalKey := skillkey.GitFromRepoURLOrEmpty(canonicalRepoURL, subPath)
		if installedIndex.IsInstalled(filepath.Base(skillPath), logicalKey) {
			continue
		}
		sk, err := a.storage.Import(skillPath, category, skill.SourceGitHub, canonicalRepoURL, subPath)
		if err == skill.ErrSkillExists {
			continue
		}
		if err != nil {
			return err
		}
		sha, _ := coregit.GetSubPathSHA(a.ctx, repoLocalDir, subPath)
		sk.SourceSHA = sha
		_ = a.storage.UpdateMeta(sk)
		imported = append(imported, sk)
	}
	a.autoPushImportedSkillsToAgents("starred.import", imported)
	a.scheduleAutoBackup()
	return nil
}
