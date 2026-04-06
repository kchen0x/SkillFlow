package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
	"time"

	"github.com/shinerio/skillflow/core/config"
	daemonruntime "github.com/shinerio/skillflow/core/platform/daemon"
	"github.com/shinerio/skillflow/core/platform/logging"
)

const helperNotifyTimeout = 3 * time.Second

var (
	startDaemonServiceFn = daemonruntime.StartService
	daemonServicePathFn  = daemonServicePath
	sendUIControlCommandFn = func(command string) error {
		return sendLoopbackControlCommand(uiControlPath(), command)
	}
)

type helperController struct {
	loggerMu     sync.Mutex
	logger       *logging.Logger
	helperServer *loopbackControlServer
	daemonService *daemonruntime.Service
	daemonApp    *App
	quitCh       chan struct{}
	quitOnce     sync.Once
	uiArgs       []string
	uiMu         sync.Mutex
	uiCmd        *exec.Cmd
}

func bootstrapHelperProcess(uiArgs []string) error {
	if err := sendLoopbackControlCommand(helperControlPath(), controlCommandShowUI); err == nil {
		return nil
	}

	lock, err := acquireHelperInstanceLock()
	if errors.Is(err, errHelperAlreadyRunning) {
		return waitForHelperShowUI(helperNotifyTimeout)
	}
	if err != nil {
		return err
	}
	defer func() {
		if releaseErr := lock.Release(); releaseErr != nil {
			fmt.Printf("helper lock release failed: %v\n", releaseErr)
		}
	}()

	helper := newHelperController(uiArgs)
	return helper.run()
}

func newHelperController(uiArgs []string) *helperController {
	return &helperController{
		logger: loadHelperLogger(),
		quitCh: make(chan struct{}),
		uiArgs: append([]string(nil), uiArgs...),
	}
}

func (h *helperController) run() error {
	h.logInfof("helper process started, platform=%s", goruntime.GOOS)

	if err := prepareHelperRuntime(); err != nil {
		return err
	}
	if err := h.initializeDaemonBackend(); err != nil {
		return err
	}

	server, err := startLoopbackControlServer(helperControlPath(), h.handleHelperControlCommand)
	if err != nil {
		return err
	}
	h.helperServer = server
	defer h.closeHelperServer()
	defer h.closeDaemonService()

	if err := setupTray(h); err != nil {
		return err
	}
	defer teardownTray()

	if err := h.ensureUIRunningAndFocused(); err != nil {
		h.logErrorf("helper initial ui launch failed: %v", err)
	}

	if err := runHelperEventLoop(h.quitCh); err != nil {
		return err
	}

	h.logInfof("helper process stopped")
	return nil
}

func (h *helperController) handleHelperControlCommand(command string) error {
	switch command {
	case controlCommandShowUI:
		return h.ensureUIRunningAndFocused()
	case controlCommandQuit:
		h.quitApp()
		return nil
	default:
		return fmt.Errorf("unsupported helper control command: %s", command)
	}
}

func (h *helperController) showMainWindow() {
	if err := h.ensureUIRunningAndFocused(); err != nil {
		h.logErrorf("helper show ui failed: %v", err)
	}
}

func (h *helperController) hideMainWindow() {
	if err := h.stopUI(); err != nil && !isControlEndpointMissing(err) {
		h.logErrorf("helper hide ui failed: %v", err)
	}
}

func (h *helperController) quitApp() {
	h.quitOnce.Do(func() {
		h.logInfof("helper quit started")
		if err := h.stopUI(); err != nil {
			h.logErrorf("helper stop ui failed: %v", err)
		}
		h.closeHelperServer()
		teardownTray()
		stopHelperEventLoop()
		close(h.quitCh)
		h.logInfof("helper quit completed")
	})
}

func (h *helperController) ensureUIRunningAndFocused() error {
	h.uiMu.Lock()
	defer h.uiMu.Unlock()

	if err := sendUIControlCommandFn(controlCommandShow); err == nil {
		return nil
	}

	_ = os.Remove(uiControlPath())
	if h.uiCmd == nil {
		if err := h.launchUIProcessLocked(); err != nil {
			return err
		}
	}

	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		if err := sendUIControlCommandFn(controlCommandShow); err == nil {
			return nil
		}
		time.Sleep(150 * time.Millisecond)
	}
	return fmt.Errorf("ui process did not become ready")
}

func (h *helperController) launchUIProcessLocked() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	args := append([]string{internalUIFlag}, h.uiArgs...)
	cmd := exec.Command(exePath, args...)
	cmd.Dir = filepath.Dir(exePath)
	if err := cmd.Start(); err != nil {
		return err
	}
	h.uiCmd = cmd
	h.logInfof("helper ui launch started, pid=%d", cmd.Process.Pid)

	go func(cmd *exec.Cmd) {
		_ = cmd.Wait()
		h.uiMu.Lock()
		if h.uiCmd == cmd {
			h.uiCmd = nil
		}
		h.uiMu.Unlock()
	}(cmd)

	return nil
}

func (h *helperController) stopUI() error {
	h.uiMu.Lock()
	cmd := h.uiCmd
	h.uiMu.Unlock()

	err := sendUIControlCommandFn(controlCommandQuit)
	if err == nil {
		h.clearTrackedUIProcess(cmd)
		return nil
	}

	if cmd != nil && cmd.Process != nil {
		if killErr := cmd.Process.Kill(); killErr != nil {
			return killErr
		}
		h.clearTrackedUIProcess(cmd)
		return nil
	}
	if isControlEndpointMissing(err) {
		h.clearTrackedUIProcess(cmd)
		return nil
	}
	return err
}

func (h *helperController) clearTrackedUIProcess(cmd *exec.Cmd) {
	h.uiMu.Lock()
	defer h.uiMu.Unlock()
	if h.uiCmd == cmd {
		h.uiCmd = nil
	}
}

func (h *helperController) closeHelperServer() {
	if h.helperServer == nil {
		return
	}
	if err := h.helperServer.Close(); err != nil {
		h.logErrorf("helper control server stop failed: %v", err)
	}
	h.helperServer = nil
}

func (h *helperController) closeDaemonService() {
	if h.daemonService == nil {
		return
	}
	if err := h.daemonService.Close(); err != nil {
		h.logErrorf("daemon service stop failed: %v", err)
	}
	h.daemonService = nil
}

func (h *helperController) initializeDaemonBackend() error {
	app := newDaemonAppFn()
	app.ctx = context.Background()

	rt, err := newDaemonRuntimeFn(appDataDirFunc(), daemonruntime.Dependencies{
		RunUpgrade:          runStartupUpgrade,
		LoadConfig:          loadStartupConfig,
		NewLogger:           logging.New,
		SyncLaunchAtLogin:   app.syncLaunchAtLogin,
		RegisterAdapters:    registerAdapters,
		RegisterProviders:   registerProviders,
		BuiltinStarredRepos: builtinStarredRepoURLs,
	})
	if err != nil {
		return err
	}

	app.applyRuntime(rt)
	h.daemonApp = app
	service, err := startDaemonServiceFn(daemonServicePathFn(), daemonServiceHandlers(app))
	if err != nil {
		return err
	}
	h.daemonService = service
	startAppAutoSyncTimerFn(app, rt.ConfigSnapshot.Cloud.SyncIntervalMinutes)
	startAppBackgroundTasksFn(app)
	return nil
}

func daemonServiceHandlers(app *App) map[string]daemonruntime.ServiceHandler {
	return map[string]daemonruntime.ServiceHandler{
		"GetConfig": func(ctx context.Context, params json.RawMessage) (any, error) {
			return app.GetConfig()
		},
		"ListCloudProviders": func(ctx context.Context, params json.RawMessage) (any, error) {
			return app.ListCloudProviders(), nil
		},
		"GetAppVersion": func(ctx context.Context, params json.RawMessage) (any, error) {
			return app.GetAppVersion(), nil
		},
	}
}

func (h *helperController) logDebugf(format string, args ...any) {
	h.logf(logging.LevelDebug, format, args...)
}

func (h *helperController) logInfof(format string, args ...any) {
	h.logf(logging.LevelInfo, format, args...)
}

func (h *helperController) logErrorf(format string, args ...any) {
	h.logf(logging.LevelError, format, args...)
}

func (h *helperController) logf(level logging.Level, format string, args ...any) {
	h.loggerMu.Lock()
	logger := h.logger
	h.loggerMu.Unlock()
	if logger == nil {
		return
	}
	logger.Logf(level, format, args...)
}

func waitForHelperShowUI(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		lastErr = sendLoopbackControlCommand(helperControlPath(), controlCommandShowUI)
		if lastErr == nil {
			return nil
		}
		time.Sleep(150 * time.Millisecond)
	}
	if lastErr == nil {
		lastErr = errHelperAlreadyRunning
	}
	return lastErr
}

func loadHelperLogger() *logging.Logger {
	dataDir := config.AppDataDir()
	cfgService := config.NewService(dataDir)
	cfg, err := cfgService.Load()
	if err != nil {
		cfg = config.DefaultConfig(dataDir)
	}
	logger, err := logging.New(filepath.Join(dataDir, "logs"), cfg.LogLevel)
	if err != nil {
		return nil
	}
	return logger
}

func isControlEndpointMissing(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, os.ErrNotExist) {
		return true
	}
	text := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(text, "no such file") || strings.Contains(text, "connection refused")
}
