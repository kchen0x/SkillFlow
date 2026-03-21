package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
	"time"

	agentapp "github.com/shinerio/skillflow/core/agentintegration/app"
	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
	backupdomain "github.com/shinerio/skillflow/core/backup/domain"
	"github.com/shinerio/skillflow/core/config"
	"github.com/shinerio/skillflow/core/platform/eventbus"
	platformgit "github.com/shinerio/skillflow/core/platform/git"
	"github.com/shinerio/skillflow/core/platform/logging"
	skillcatalogapp "github.com/shinerio/skillflow/core/skillcatalog/app"
	skilldomain "github.com/shinerio/skillflow/core/skillcatalog/domain"
	skillrepo "github.com/shinerio/skillflow/core/skillcatalog/infra/repository"
	"github.com/shinerio/skillflow/core/skillkey"
	sourcedomain "github.com/shinerio/skillflow/core/skillsource/domain"
	sourcerepo "github.com/shinerio/skillflow/core/skillsource/infra/repository"
	"github.com/shinerio/skillflow/core/upgrade"
	"github.com/shinerio/skillflow/core/viewstate"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx                context.Context
	hub                *eventbus.Hub
	sysLog             *logging.Logger
	storage            *skillcatalogapp.Service
	config             *config.Service
	starStorage        *sourcerepo.StarRepoStorage
	cacheDir           string
	viewCache          *viewstate.Manager
	promptImports      *promptImportSessionStore
	startupOnce        sync.Once
	initialWindowState config.WindowState
	autostartFactory   func() (launchAtLoginController, error)

	// Git sync state
	gitConflictMu      sync.Mutex
	gitConflictPending bool
	stopAutoSync       chan struct{}

	backupResultMu       sync.RWMutex
	lastBackupResult     []backupdomain.RemoteFile
	lastBackupAt         time.Time
	windowVisibilityMu   sync.Mutex
	windowVisibilityInit bool
	windowVisible        bool
	uiControlMu          sync.Mutex
	uiControlServer      *loopbackControlServer
}

const defaultCategoryName = "Default"
const proxyResponseHeaderTimeout = 10 * time.Second

var builtinStarredRepoURLs = []string{
	"https://github.com/anthropics/skills.git",
	"https://github.com/ComposioHQ/awesome-claude-skills.git",
	"https://github.com/affaan-m/everything-claude-code.git",
}

var appDataDirFunc = config.AppDataDir

var runStartupUpgrade = upgrade.Run

var cloneOrUpdateRepo = platformgit.CloneOrUpdate

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
		hub:           eventbus.NewHub(),
		promptImports: newPromptImportSessionStore(),
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
	a.storage = skillcatalogapp.NewService(skillrepo.NewFilesystemStorage(cfg.SkillsStorageDir))
	a.cacheDir = filepath.Join(dataDir, "cache")
	a.viewCache = viewstate.NewManager(filepath.Join(a.cacheDir, "viewstate"))
	a.starStorage = sourcerepo.NewStarRepoStorageWithBuiltins(filepath.Join(dataDir, "star_repos.json"), builtinStarredRepoURLs)
	registerAdapters()
	registerProviders()
	if ctx != nil {
		go forwardEvents(ctx, a.hub)
	}
	a.logDebugf("startup background tasks deferred until ui ready")
	a.startAutoSyncTimer(cfg.Cloud.SyncIntervalMinutes)
}

// proxyHTTPClient builds an *http.Client configured according to the saved proxy settings.
// It preserves Go's default transport behavior and bounds header waits so proxy/network
// stalls do not hang forever.
func (a *App) proxyHTTPClient() *http.Client {
	if a.config == nil {
		return proxyHTTPClientWithConfig(config.ProxyConfig{Mode: config.ProxyModeSystem})
	}

	cfg, err := a.config.Load()
	if err != nil {
		return proxyHTTPClientWithConfig(config.ProxyConfig{Mode: config.ProxyModeSystem})
	}
	return proxyHTTPClientWithConfig(cfg.Proxy)
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
	profile := a.backupProfile(cfg)
	isGit := profile.Provider == backupdomain.GitProviderName
	if isGit {
		a.hub.Publish(eventbus.Event{Type: eventbus.EventGitSyncStarted})
	}
	a.hub.Publish(eventbus.Event{Type: eventbus.EventBackupStarted})
	result, err := a.newBackupService().RunBackup(a.ctx, profile,
		func(file string) {
			a.hub.Publish(eventbus.Event{
				Type:    eventbus.EventBackupProgress,
				Payload: eventbus.BackupProgressPayload{CurrentFile: file},
			})
		})
	if err != nil {
		a.logErrorf("backup failed: %v", err)
		var conflictErr *backupdomain.GitConflictError
		if isGit && errors.As(err, &conflictErr) {
			a.publishGitConflict(conflictErr)
		}
		if isGit {
			a.hub.Publish(eventbus.Event{Type: eventbus.EventGitSyncFailed, Payload: err.Error()})
		}
		a.hub.Publish(eventbus.Event{Type: eventbus.EventBackupFailed, Payload: err.Error()})
		return err
	} else {
		a.recordBackupResult(result.Files, result.Snapshot)
		payload := eventbus.BackupCompletedPayload{Files: result.Files, CompletedAt: a.GetLastBackupCompletedAt()}
		a.logInfof("backup completed")
		if isGit {
			a.clearGitConflictPending()
			a.hub.Publish(eventbus.Event{Type: eventbus.EventGitSyncCompleted, Payload: payload})
		}
		a.hub.Publish(eventbus.Event{Type: eventbus.EventBackupCompleted, Payload: payload})
		return nil
	}
}

func (a *App) publishGitConflict(conflictErr *backupdomain.GitConflictError) {
	a.gitConflictMu.Lock()
	a.gitConflictPending = true
	a.gitConflictMu.Unlock()
	a.hub.Publish(eventbus.Event{
		Type: eventbus.EventGitConflict,
		Payload: eventbus.GitConflictPayload{
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
	a.storage = skillcatalogapp.NewService(skillrepo.NewFilesystemStorage(cfg.SkillsStorageDir))
	a.startAutoSyncTimer(cfg.Cloud.SyncIntervalMinutes)
}

// gitPullOnStartup pulls from the remote git repo at startup when the git provider is enabled.
func (a *App) gitPullOnStartup() {
	cfg, err := a.config.Load()
	if err != nil || !cfg.Cloud.Enabled || cfg.Cloud.Provider != backupdomain.GitProviderName {
		return
	}
	beforeRestore := a.captureCloudRestoreState()
	a.logInfof("startup git pull started")
	a.hub.Publish(eventbus.Event{Type: eventbus.EventGitSyncStarted})
	result, err := a.newBackupService().RestoreBackup(a.ctx, a.backupProfile(cfg))
	if err != nil {
		a.logErrorf("startup git pull failed: %v", err)
		var conflictErr *backupdomain.GitConflictError
		if errors.As(err, &conflictErr) {
			a.publishGitConflict(conflictErr)
			a.hub.Publish(eventbus.Event{Type: eventbus.EventGitSyncFailed, Payload: err.Error()})
		} else {
			a.hub.Publish(eventbus.Event{Type: eventbus.EventGitSyncFailed, Payload: err.Error()})
		}
		return
	}
	a.recordBackupResult(result.Files, result.Snapshot)
	if err := a.handleRestoredCloudState(beforeRestore, "startup.git.pull"); err != nil {
		a.logErrorf("startup git pull failed: restore compensation failed: %v", err)
		a.hub.Publish(eventbus.Event{Type: eventbus.EventGitSyncFailed, Payload: err.Error()})
		return
	}
	a.logInfof("startup git pull completed")
	a.clearGitConflictPending()
	a.hub.Publish(eventbus.Event{Type: eventbus.EventGitSyncCompleted, Payload: eventbus.BackupCompletedPayload{Files: result.Files, CompletedAt: a.GetLastBackupCompletedAt()}})
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
	cfg, err := a.config.Load()
	if err != nil {
		a.logErrorf("git conflict resolution failed: strategy=%s load config failed: %v", action, err)
		return err
	}
	backupDir := a.newBackupService().BackupRootDir(a.backupProfile(cfg))
	a.logInfof("git conflict resolution started: strategy=%s, backupDir=%s", action, backupDir)
	result, err := a.newBackupService().ResolveGitConflict(a.backupProfile(cfg), useLocal)
	if err != nil {
		a.logErrorf("git conflict resolution failed: strategy=%s, backupDir=%s, err=%v", action, backupDir, err)
		return err
	}
	a.gitConflictMu.Lock()
	a.gitConflictPending = false
	a.gitConflictMu.Unlock()
	a.recordBackupResult(result.Files, result.Snapshot)
	a.reloadStateFromDisk()
	a.hub.Publish(eventbus.Event{Type: eventbus.EventGitSyncCompleted, Payload: eventbus.BackupCompletedPayload{Files: result.Files, CompletedAt: a.GetLastBackupCompletedAt()}})
	a.logInfof("git conflict resolution completed: strategy=%s, backupDir=%s", action, backupDir)
	return nil
}

func (a *App) recordBackupResult(files []backupdomain.RemoteFile, _ backupdomain.Snapshot) {
	a.setLastBackupResult(files)
}

func (a *App) setLastBackupResult(files []backupdomain.RemoteFile) {
	copied := make([]backupdomain.RemoteFile, len(files))
	copy(copied, files)

	a.backupResultMu.Lock()
	a.lastBackupResult = copied
	a.lastBackupAt = time.Now().UTC()
	a.backupResultMu.Unlock()
}

func (a *App) GetLastBackupChanges() []backupdomain.RemoteFile {
	a.backupResultMu.RLock()
	defer a.backupResultMu.RUnlock()

	copied := make([]backupdomain.RemoteFile, len(a.lastBackupResult))
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
		if errors.Is(err, skillrepo.ErrCategoryNotEmpty) {
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

func (a *App) GetSkillMeta(skillID string) (*skilldomain.SkillMeta, error) {
	sk, err := a.storage.Get(skillID)
	if err != nil {
		return nil, err
	}
	return skilldomain.ReadMeta(sk.Path)
}

// GetSkillMetaByPath reads skill.md frontmatter from a skill directory path (no ID required).
func (a *App) GetSkillMetaByPath(path string) (*skilldomain.SkillMeta, error) {
	return skilldomain.ReadMeta(path)
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
		if canonical, err := platformgit.CanonicalRepoURL(rawURL); err == nil {
			target = canonical
		}
	}
	runtime.BrowserOpenURL(a.ctx, target)
	return nil
}

// PushStarSkillsToAgents copies starred skill directories directly to the push directory of each
// specified agent, skipping skills that already exist. Returns structured conflicts.
func (a *App) PushStarSkillsToAgents(skillPaths []string, agentNames []string) ([]agentdomain.PushConflict, error) {
	a.logInfof("push starred skills to agents started: skillCount=%d agentCount=%d", len(skillPaths), len(agentNames))
	cfg, _ := a.config.Load()
	tempSkills := make([]*skilldomain.InstalledSkill, 0, len(skillPaths))
	for _, skillPath := range skillPaths {
		tempSkills = append(tempSkills, &skilldomain.InstalledSkill{
			Name: filepath.Base(skillPath),
			Path: skillPath,
		})
	}
	conflicts, err := newAgentIntegrationService().PushSkills(a.ctx, agentProfiles(cfg.Agents), agentNames, tempSkills, false)
	if err != nil {
		a.logErrorf("push starred skills to agents failed: err=%v", err)
		return nil, err
	}
	a.logInfof("push starred skills to agents completed: conflicts=%d", len(conflicts))
	return conflicts, nil
}

// PushStarSkillsToAgentsForce copies starred skill directories to agent push directories,
// overwriting any existing skills.
func (a *App) PushStarSkillsToAgentsForce(skillPaths []string, agentNames []string) error {
	a.logInfof("force push starred skills to agents started: skillCount=%d agentCount=%d", len(skillPaths), len(agentNames))
	cfg, _ := a.config.Load()
	tempSkills := make([]*skilldomain.InstalledSkill, 0, len(skillPaths))
	for _, skillPath := range skillPaths {
		tempSkills = append(tempSkills, &skilldomain.InstalledSkill{
			Name: filepath.Base(skillPath),
			Path: skillPath,
		})
	}
	if _, err := newAgentIntegrationService().PushSkills(a.ctx, agentProfiles(cfg.Agents), agentNames, tempSkills, true); err != nil {
		a.logErrorf("force push starred skills to agents failed: err=%v", err)
		return err
	}
	a.logInfof("force push starred skills to agents completed")
	return nil
}

func (a *App) ImportLocal(dir, category string) (*skilldomain.InstalledSkill, error) {
	category = normalizeCategoryName(category)
	sk, err := a.storage.Import(dir, category, skilldomain.SourceManual, "", "")
	if err != nil {
		return nil, err
	}
	a.autoPushImportedSkillsToAgents("local.import", []*skilldomain.InstalledSkill{sk})
	a.scheduleAutoBackup()
	return sk, nil
}

func (a *App) autoPushAgentTargets(cfg config.AppConfig) []agentdomain.AgentProfile {
	if len(cfg.AutoPushAgents) == 0 {
		return nil
	}
	selected := make(map[string]struct{}, len(cfg.AutoPushAgents))
	for _, name := range cfg.AutoPushAgents {
		selected[name] = struct{}{}
	}
	targets := make([]agentdomain.AgentProfile, 0, len(selected))
	for _, agent := range cfg.Agents {
		if !agent.Enabled {
			continue
		}
		if _, ok := selected[agent.Name]; ok {
			targets = append(targets, agentProfile(agent))
		}
	}
	return targets
}

func (a *App) autoPushImportedSkillsToAgents(source string, imported []*skilldomain.InstalledSkill) {
	a.autoPushSkillsToConfiguredAgents(source, imported, false)
}

func (a *App) autoPushSkillsToConfiguredAgents(source string, skills []*skilldomain.InstalledSkill, overwriteExisting bool) {
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
	targetNames := make([]string, 0, len(targets))
	for _, profile := range targets {
		targetNames = append(targetNames, profile.Name)
	}
	conflicts, err := newAgentIntegrationService().PushSkills(a.ctx, targets, targetNames, skills, overwriteExisting)
	if err != nil {
		a.logErrorf("auto push imported skills failed: source=%s err=%v", source, err)
		return
	}
	a.logInfof("auto push imported skills completed: source=%s skillCount=%d agentCount=%d conflicts=%d overwriteExisting=%t", source, len(skills), len(targets), len(conflicts), overwriteExisting)
}

// --- Sync ---

func (a *App) GetEnabledAgents() ([]config.AgentConfig, error) {
	cfg, err := a.config.Load()
	if err != nil {
		return nil, err
	}
	return agentConfigs(newAgentIntegrationService().EnabledProfiles(agentProfiles(cfg.Agents))), nil
}

// ScanAgentSkills lists all skills in an agent's configured scan directories for the pull page.
func (a *App) ScanAgentSkills(agentName string) ([]agentdomain.AgentSkillCandidate, error) {
	cfg, _ := a.config.Load()
	profile, ok := agentapp.FindProfile(agentProfiles(cfg.Agents), agentName)
	if !ok {
		return nil, nil
	}
	_, installedIndex, err := a.installedIndex()
	if err != nil {
		return nil, err
	}
	return newAgentIntegrationService().ScanAgentSkills(a.ctx, profile, installedIndex, a.buildAgentPresenceIndex(installedIndex), a.repoScanMaxDepth())
}

// ListAgentSkills returns all skills for an agent, annotated with whether each
// skill lives in the push directory and/or the scan directories.
func (a *App) ListAgentSkills(agentName string) ([]agentdomain.AgentSkillEntry, error) {
	cfg, err := a.config.Load()
	if err != nil {
		return nil, err
	}
	profile, ok := agentapp.FindProfile(agentProfiles(cfg.Agents), agentName)
	if !ok {
		return nil, fmt.Errorf("agent %s not found", agentName)
	}
	_, installedIndex, err := a.installedIndex()
	if err != nil {
		return nil, err
	}
	return newAgentIntegrationService().ListAgentSkills(a.ctx, profile, installedIndex, a.buildAgentPresenceIndex(installedIndex), a.repoScanMaxDepth())
}

// DeleteAgentSkill removes a skill directory from an agent's push directory.
// Returns an error if skillPath is not within the agent's configured push directory.
func (a *App) DeleteAgentSkill(agentName string, skillPath string) error {
	cfg, err := a.config.Load()
	if err != nil {
		return err
	}
	profile, ok := agentapp.FindProfile(agentProfiles(cfg.Agents), agentName)
	if !ok {
		return fmt.Errorf("agent %s not found", agentName)
	}
	a.logInfof("DeleteAgentSkill: deleting %s from agent %s push dir started", filepath.Base(skillPath), agentName)
	if err := newAgentIntegrationService().DeletePushedSkill(profile, skillPath); err != nil {
		a.logErrorf("DeleteAgentSkill: delete %s failed: %v", skillPath, err)
		return err
	}
	a.logInfof("DeleteAgentSkill: deleted %s from agent %s push dir completed", filepath.Base(skillPath), agentName)
	return nil
}

// CheckMissingAgentPushDirs returns agent names and paths whose push directory does not yet exist.
// Each element is map{"name": agentName, "dir": pushDir}.
func (a *App) CheckMissingAgentPushDirs(agentNames []string) ([]agentdomain.MissingPushDir, error) {
	cfg, _ := a.config.Load()
	return newAgentIntegrationService().CheckMissingPushDirs(agentProfiles(cfg.Agents), agentNames)
}

// PushToAgents pushes selected skills to target agents.
// Returns structured conflicts that were skipped.
func (a *App) PushToAgents(skillIDs []string, agentNames []string) ([]agentdomain.PushConflict, error) {
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
	var selected []*skilldomain.InstalledSkill
	for _, sk := range skills {
		if idSet[sk.ID] {
			selected = append(selected, sk)
		}
	}
	conflicts, err := newAgentIntegrationService().PushSkills(a.ctx, agentProfiles(cfg.Agents), agentNames, selected, false)
	if err != nil {
		a.logErrorf("push skills to agents failed: %v", err)
		return nil, err
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
	var selected []*skilldomain.InstalledSkill
	for _, sk := range skills {
		if idSet[sk.ID] {
			selected = append(selected, sk)
		}
	}
	if _, err := newAgentIntegrationService().PushSkills(a.ctx, agentProfiles(cfg.Agents), agentNames, selected, true); err != nil {
		a.logErrorf("force push skills to agents failed: %v", err)
		return err
	}
	a.logInfof("force push skills to agents completed")
	return nil
}

// PullFromAgent imports selected skills from an agent into SkillFlow storage.
func (a *App) PullFromAgent(agentName string, skillPaths []string, category string) ([]string, error) {
	category = normalizeCategoryName(category)
	cfg, _ := a.config.Load()
	profile, ok := agentapp.FindProfile(agentProfiles(cfg.Agents), agentName)
	if !ok {
		return nil, nil
	}
	_, installedIndex, err := a.installedIndex()
	if err != nil {
		return nil, err
	}
	scanned, err := newAgentIntegrationService().ScanAgentSkills(a.ctx, profile, installedIndex, a.buildAgentPresenceIndex(installedIndex), a.repoScanMaxDepth())
	if err != nil {
		return nil, err
	}
	candidates := agentapp.SelectAgentSkillCandidates(scanned, skillPaths)
	var conflicts []string
	imported := make([]*skilldomain.InstalledSkill, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.Imported {
			conflicts = append(conflicts, candidate.Path)
			continue
		}
		sk, err := a.storage.Import(candidate.Path, category, skilldomain.SourceManual, "", "")
		if err == skillrepo.ErrSkillExists {
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

// PullFromAgentForce imports selected skills, overwriting existing ones.
func (a *App) PullFromAgentForce(agentName string, skillPaths []string, category string) error {
	category = normalizeCategoryName(category)
	cfg, _ := a.config.Load()
	profile, ok := agentapp.FindProfile(agentProfiles(cfg.Agents), agentName)
	if !ok {
		return nil
	}
	_, installedIndex, err := a.installedIndex()
	if err != nil {
		return err
	}
	scanned, err := newAgentIntegrationService().ScanAgentSkills(a.ctx, profile, installedIndex, a.buildAgentPresenceIndex(installedIndex), a.repoScanMaxDepth())
	if err != nil {
		return err
	}
	imported := make([]*skilldomain.InstalledSkill, 0, len(skillPaths))
	for _, candidate := range agentapp.SelectAgentSkillCandidates(scanned, skillPaths) {
		existing, _ := a.storage.ListAll()
		for _, e := range existing {
			logicalKey, logicalErr := skilldomain.LogicalKey(e)
			if (candidate.LogicalKey != "" && logicalErr == nil && logicalKey == candidate.LogicalKey) || (candidate.LogicalKey == "" && e.Name == candidate.Name) {
				_ = a.storage.Delete(e.ID)
			}
		}
		sk, err := a.storage.Import(candidate.Path, category, skilldomain.SourceManual, "", "")
		if err != nil {
			return err
		}
		imported = append(imported, sk)
	}
	a.autoPushImportedSkillsToAgents("agent.pull.force", imported)
	a.scheduleAutoBackup()
	return nil
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

func (a *App) ListCloudFiles() ([]backupdomain.RemoteFile, error) {
	cfg, err := a.config.Load()
	if err != nil {
		a.logErrorf("list cloud files failed: load config failed: %v", err)
		return nil, err
	}
	a.logInfof("list cloud files started (provider=%s, remotePath=%s)", cfg.Cloud.Provider, cfg.Cloud.RemotePath)
	files, err := a.newBackupService().ListRemoteBackupFiles(a.ctx, a.backupProfile(cfg))
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
	profile := a.backupProfile(cfg)
	isGit := profile.Provider == backupdomain.GitProviderName
	if isGit {
		a.hub.Publish(eventbus.Event{Type: eventbus.EventGitSyncStarted})
	}
	result, err := a.newBackupService().RestoreBackup(a.ctx, profile)
	if err != nil {
		a.logErrorf("restore from cloud failed: %v", err)
		var conflictErr *backupdomain.GitConflictError
		if isGit && errors.As(err, &conflictErr) {
			a.publishGitConflict(conflictErr)
		}
		if isGit {
			a.hub.Publish(eventbus.Event{Type: eventbus.EventGitSyncFailed, Payload: err.Error()})
		}
		return err
	}
	a.recordBackupResult(result.Files, result.Snapshot)
	a.logInfof("restore from cloud completed")
	if err := a.handleRestoredCloudState(beforeRestore, "cloud.restore"); err != nil {
		a.logErrorf("restore from cloud failed: restore compensation failed: %v", err)
		if isGit {
			a.hub.Publish(eventbus.Event{Type: eventbus.EventGitSyncFailed, Payload: err.Error()})
		}
		return err
	}
	if isGit {
		a.clearGitConflictPending()
		a.hub.Publish(eventbus.Event{Type: eventbus.EventGitSyncCompleted, Payload: eventbus.BackupCompletedPayload{Files: result.Files, CompletedAt: a.GetLastBackupCompletedAt()}})
	}
	return nil
}

// ListCloudProviders returns all registered provider names and their required credential fields.
func (a *App) ListCloudProviders() []map[string]any {
	var result []map[string]any
	for _, p := range allCloudProviders() {
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
		Skills     []*skilldomain.InstalledSkill
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
				a.hub.Publish(eventbus.Event{
					Type: eventbus.EventUpdateAvailable,
					Payload: eventbus.UpdateAvailablePayload{
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
	a.autoPushSkillsToConfiguredAgents("skill.update", []*skilldomain.InstalledSkill{sk}, true)
	a.scheduleAutoBackup()
	a.hub.Publish(eventbus.Event{Type: eventbus.EventSkillsUpdated})
	a.logInfof("update skill completed: id=%s name=%s", skillID, sk.Name)
	return nil
}

func (a *App) autoUpdateInstalledSkillsForRepos(source string, repoURLs []string) {
	cfg, err := a.config.Load()
	if err != nil {
		a.logErrorf("auto update installed skills failed: source=%s load config failed: %v", source, err)
		return
	}
	if !cfg.AutoUpdateSkills {
		a.logDebugf("auto update installed skills skipped: source=%s reason=disabled", source)
		return
	}

	repoSources := make(map[string]struct{}, len(repoURLs))
	for _, repoURL := range repoURLs {
		repoSource, err := platformgit.RepoSource(repoURL)
		if err != nil || strings.TrimSpace(repoSource) == "" {
			if err != nil {
				a.logErrorf("auto update installed skills failed: source=%s repo=%s err=%v", source, repoURL, err)
			}
			continue
		}
		repoSources[repoSource] = struct{}{}
	}
	if len(repoSources) == 0 {
		a.logDebugf("auto update installed skills skipped: source=%s reason=no-repos", source)
		return
	}

	skills, err := a.storage.ListAll()
	if err != nil {
		a.logErrorf("auto update installed skills failed: source=%s load skills failed: %v", source, err)
		return
	}

	updatedCount := 0
	for _, sk := range skills {
		if sk == nil || !sk.IsGitHub() {
			continue
		}
		repoSource, err := platformgit.RepoSource(sk.SourceURL)
		if err != nil {
			a.logErrorf("auto update installed skills failed: source=%s skillID=%s name=%s err=%v", source, sk.ID, sk.Name, err)
			continue
		}
		if _, ok := repoSources[repoSource]; !ok {
			continue
		}

		_, latestSHA, err := a.cachedSkillSourceDir(sk)
		if err != nil {
			a.logErrorf("auto update installed skills failed: source=%s skillID=%s name=%s err=%v", source, sk.ID, sk.Name, err)
			continue
		}
		if latestSHA == "" || latestSHA == sk.SourceSHA {
			continue
		}

		a.logInfof("auto update installed skill started: source=%s skillID=%s name=%s repo=%s", source, sk.ID, sk.Name, sk.SourceURL)
		if err := a.UpdateSkill(sk.ID); err != nil {
			a.logErrorf("auto update installed skills failed: source=%s skillID=%s name=%s err=%v", source, sk.ID, sk.Name, err)
			continue
		}
		updatedCount++
		a.logInfof("auto update installed skill completed: source=%s skillID=%s name=%s repo=%s", source, sk.ID, sk.Name, sk.SourceURL)
	}

	a.logInfof("auto update installed skills completed: source=%s updated=%d", source, updatedCount)
}

func (a *App) cachedSkillSourceDir(sk *skilldomain.InstalledSkill) (string, string, error) {
	if sk == nil {
		return "", "", fmt.Errorf("skill is required")
	}
	if a.config == nil {
		return "", "", fmt.Errorf("config service is not initialized")
	}

	cacheDir, err := platformgit.CacheDir(a.config.DataDir(), sk.SourceURL)
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
	latestSHA, err := platformgit.GetSubPathSHA(ctx, cacheDir, shaPath)
	if err != nil {
		return "", "", err
	}
	return sourceDir, latestSHA, nil
}

func (a *App) refreshUpdatedPushedSkillCopies(sk *skilldomain.InstalledSkill) error {
	if sk == nil {
		return nil
	}

	cfg, err := a.config.Load()
	if err != nil {
		a.logErrorf("refresh updated pushed skill copies failed: skill=%s load config failed: %v", sk.Name, err)
		return err
	}

	a.logInfof("refresh updated pushed skill copies started: skill=%s", sk.Name)
	if err := newAgentIntegrationService().RefreshPushedCopies(a.ctx, agentProfiles(cfg.Agents), sk); err != nil {
		a.logErrorf("refresh updated pushed skill copies failed: skill=%s err=%v", sk.Name, err)
		return err
	}
	a.logInfof("refresh updated pushed skill copies completed: skill=%s", sk.Name)
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

func (a *App) AddStarredRepo(repoURL string) (*sourcedomain.StarRepo, error) {
	a.logInfof("add starred repo requested: %s", repoURL)
	repo, err := a.newSkillsourceService().TrackStarRepo(a.cloneContext(), filepath.Dir(a.cacheDir), repoURL, a.gitProxyURL())
	if err != nil {
		if platformgit.IsSSHAuthError(err) {
			a.logErrorf("add starred repo failed: %v", err)
			return nil, fmt.Errorf("AUTH_SSH:%s", err.Error())
		}
		if platformgit.IsAuthError(err) {
			a.logErrorf("add starred repo failed: %v", err)
			return nil, fmt.Errorf("AUTH_HTTP:%s", err.Error())
		}
		a.logErrorf("add starred repo failed: %v", err)
		return nil, err
	}
	a.logInfof("add starred repo completed: %s", repoURL)
	return repo, nil
}

// AddStarredRepoWithCredentials clones a repo using the provided HTTP username/password,
// removing any previously failed entry for the same URL first.
func (a *App) AddStarredRepoWithCredentials(repoURL, username, password string) (*sourcedomain.StarRepo, error) {
	a.logInfof("add starred repo with credentials requested: %s", repoURL)
	repo, err := a.newSkillsourceService().TrackStarRepoWithCredentials(a.cloneContext(), filepath.Dir(a.cacheDir), repoURL, a.gitProxyURL(), username, password)
	if err != nil {
		a.logErrorf("add starred repo with credentials failed: %v", err)
		return nil, err
	}
	a.logInfof("add starred repo with credentials completed: %s", repoURL)
	return repo, nil
}

func (a *App) RemoveStarredRepo(repoURL string) error {
	a.logInfof("remove starred repo requested: %s", repoURL)
	if err := a.newSkillsourceService().UntrackStarRepo(repoURL); err != nil {
		a.logErrorf("remove starred repo failed: %v", err)
		return err
	}
	a.logInfof("remove starred repo completed: %s", repoURL)
	return nil
}

func (a *App) ListStarredRepos() ([]sourcedomain.StarRepo, error) {
	return a.newSkillsourceService().ListStarRepos()
}

func (a *App) ListAllStarSkills() ([]StarSkillEntry, error) {
	return measureOperation(a, "list_all_star_skills", func() ([]StarSkillEntry, error) {
		fingerprint, fingerprintErr := a.allStarSkillsFingerprint()
		if fingerprintErr == nil {
			var cached []StarSkillEntry
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

func (a *App) listAllStarSkillsUncached() ([]StarSkillEntry, error) {
	skills, err := a.newSkillsourceService().ListAllSourceCandidates(a.repoScanMaxDepth())
	if err != nil {
		return nil, err
	}
	_, installedIndex, err := a.installedIndex()
	if err != nil {
		return nil, err
	}
	presence := a.buildAgentPresenceIndex(installedIndex)
	all := resolveSourceCandidates(skills, installedIndex, presence)
	if all == nil {
		return []StarSkillEntry{}, nil
	}
	return all, nil
}

func (a *App) ListRepoStarSkills(repoURL string) ([]StarSkillEntry, error) {
	skills, err := a.newSkillsourceService().ListSourceCandidatesByRepo(repoURL, a.repoScanMaxDepth())
	if err != nil {
		return nil, err
	}
	_, installedIndex, err := a.installedIndex()
	if err != nil {
		return nil, err
	}
	presence := a.buildAgentPresenceIndex(installedIndex)
	resolved := resolveSourceCandidates(skills, installedIndex, presence)
	if resolved == nil {
		return []StarSkillEntry{}, nil
	}
	return resolved, nil
}

func (a *App) UpdateStarredRepo(repoURL string) error {
	a.logInfof("update starred repo started: repo=%s", repoURL)
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	repo, err := a.newSkillsourceService().RefreshStarRepo(ctx, repoURL, a.gitProxyURL())
	if err != nil {
		a.logErrorf("update starred repo failed: repo=%s err=%v", repoURL, err)
		return err
	}
	if repo == nil {
		a.logErrorf("update starred repo failed: repo=%s err=starred repo not found", repoURL)
		return nil
	}
	if strings.TrimSpace(repo.SyncError) != "" {
		a.logErrorf("update starred repo failed: repo=%s err=%s", repo.URL, repo.SyncError)
	} else {
		a.logInfof("update starred repo completed: repo=%s", repo.URL)
		a.autoUpdateInstalledSkillsForRepos("starred.refresh.one", []string{repo.URL})
	}
	a.hub.Publish(eventbus.Event{
		Type: eventbus.EventStarSyncProgress,
		Payload: eventbus.StarSyncProgressPayload{
			RepoURL:   repo.URL,
			RepoName:  repo.Name,
			SyncError: repo.SyncError,
		},
	})
	return nil
}

func (a *App) UpdateAllStarredRepos() error {
	a.logInfof("update all starred repos requested")
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	results, err := a.newSkillsourceService().RefreshAllStarRepos(ctx, a.gitProxyURL())
	if err != nil {
		a.logErrorf("update all starred repos failed: %v", err)
		return err
	}
	if len(results) == 0 {
		return nil
	}
	successfulRepoURLs := make([]string, 0, len(results))
	for _, result := range results {
		repo := result.Repo
		if strings.TrimSpace(repo.SyncError) != "" {
			a.logErrorf("update starred repo failed: repo=%s err=%s", repo.URL, repo.SyncError)
		} else {
			a.logInfof("update starred repo completed: repo=%s", repo.URL)
			successfulRepoURLs = append(successfulRepoURLs, repo.URL)
		}
		a.hub.Publish(eventbus.Event{
			Type: eventbus.EventStarSyncProgress,
			Payload: eventbus.StarSyncProgressPayload{
				RepoURL:   repo.URL,
				RepoName:  repo.Name,
				SyncError: repo.SyncError,
			},
		})
	}
	a.autoUpdateInstalledSkillsForRepos("starred.refresh.all", successfulRepoURLs)
	a.hub.Publish(eventbus.Event{Type: eventbus.EventStarSyncDone})
	a.logInfof("update all starred repos completed")
	return nil
}

func (a *App) ImportStarSkills(skillPaths []string, repoURL, category string) error {
	category = normalizeCategoryName(category)
	repos, _ := a.newSkillsourceService().ListStarRepos()
	var repoLocalDir string
	canonicalRepoURL := repoURL
	if normalized, err := platformgit.CanonicalRepoURL(repoURL); err == nil {
		canonicalRepoURL = normalized
	}
	for _, r := range repos {
		if platformgit.SameRepo(r.URL, repoURL) {
			repoLocalDir = r.LocalDir
			if normalized, err := platformgit.CanonicalRepoURL(r.URL); err == nil {
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
	imported := make([]*skilldomain.InstalledSkill, 0, len(skillPaths))
	for _, skillPath := range skillPaths {
		subPath, _ := filepath.Rel(repoLocalDir, skillPath)
		subPath = filepath.ToSlash(subPath)
		logicalKey := skillkey.GitFromRepoURLOrEmpty(canonicalRepoURL, subPath)
		if installedIndex.IsInstalled(filepath.Base(skillPath), logicalKey) {
			continue
		}
		sk, err := a.storage.Import(skillPath, category, skilldomain.SourceGitHub, canonicalRepoURL, subPath)
		if err == skillrepo.ErrSkillExists {
			continue
		}
		if err != nil {
			return err
		}
		sha, _ := platformgit.GetSubPathSHA(a.ctx, repoLocalDir, subPath)
		sk.SourceSHA = sha
		_ = a.storage.UpdateMeta(sk)
		imported = append(imported, sk)
	}
	a.autoPushImportedSkillsToAgents("starred.import", imported)
	a.scheduleAutoBackup()
	return nil
}
