package main

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/config"
	daemonruntime "github.com/shinerio/skillflow/core/platform/daemon"
	"github.com/shinerio/skillflow/core/platform/eventbus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHelperControllerInitializeDaemonBackendBuildsRuntimeAndStartsServices(t *testing.T) {
	prevNewDaemonRuntimeFn := newDaemonRuntimeFn
	prevNewDaemonAppFn := newDaemonAppFn
	prevStartDaemonServiceFn := startDaemonServiceFn
	prevDaemonServicePathFn := daemonServicePathFn
	prevStartAppAutoSyncTimerFn := startAppAutoSyncTimerFn
	prevStartAppBackgroundTasksFn := startAppBackgroundTasksFn
	prevActiveProcessRole := activeProcessRole
	t.Cleanup(func() {
		newDaemonRuntimeFn = prevNewDaemonRuntimeFn
		newDaemonAppFn = prevNewDaemonAppFn
		startDaemonServiceFn = prevStartDaemonServiceFn
		daemonServicePathFn = prevDaemonServicePathFn
		startAppAutoSyncTimerFn = prevStartAppAutoSyncTimerFn
		startAppBackgroundTasksFn = prevStartAppBackgroundTasksFn
		activeProcessRole = prevActiveProcessRole
	})

	activeProcessRole = processRoleDaemon

	dataDir := t.TempDir()
	rt := &daemonruntime.Runtime{
		DataDir:         dataDir,
		ConfigService:   config.NewService(dataDir),
		ConfigSnapshot:  config.AppConfig{Cloud: config.CloudConfig{SyncIntervalMinutes: 9}},
		Hub:             eventbus.NewHub(),
	}
	newDaemonRuntimeFn = func(gotDataDir string, deps daemonruntime.Dependencies) (*daemonruntime.Runtime, error) {
		assert.NotEmpty(t, gotDataDir)
		require.NotNil(t, deps.SyncLaunchAtLogin)
		return rt, nil
	}
	newDaemonAppFn = func() *App {
		return NewApp()
	}
	daemonServicePathFn = func() string {
		return filepath.Join(dataDir, "runtime", "daemon-service.json")
	}

	autoSyncCalls := 0
	backgroundCalls := 0
	serviceCalls := 0
	startDaemonServiceFn = func(statePath string, handlers map[string]daemonruntime.ServiceHandler) (*daemonruntime.Service, error) {
		serviceCalls++
		assert.Equal(t, filepath.Join(dataDir, "runtime", "daemon-service.json"), statePath)
		require.Contains(t, handlers, "GetConfig")
		require.Contains(t, handlers, "ListSkills")
		require.Contains(t, handlers, "ListCategories")
		require.Contains(t, handlers, "GetGitConflictPending")
		require.Contains(t, handlers, "ListCloudProviders")
		return &daemonruntime.Service{}, nil
	}
	startAppAutoSyncTimerFn = func(app *App, minutes int) {
		autoSyncCalls++
		assert.Equal(t, 9, minutes)
		assert.Same(t, rt, app.backendRuntime)
	}
	startAppBackgroundTasksFn = func(app *App) {
		backgroundCalls++
		assert.Same(t, rt, app.backendRuntime)
	}

	controller := newHelperController(nil)
	require.NoError(t, controller.initializeDaemonBackend())
	require.NotNil(t, controller.daemonApp)
	assert.Same(t, rt, controller.daemonApp.backendRuntime)
	assert.IsType(t, context.Background(), controller.daemonApp.ctx)
	require.NotNil(t, controller.daemonService)
	assert.Equal(t, 1, serviceCalls)
	assert.Equal(t, 1, autoSyncCalls)
	assert.Equal(t, 1, backgroundCalls)
}

func TestHelperControllerHideMainWindowStopsUIProcess(t *testing.T) {
	prevSendUIControlCommandFn := sendUIControlCommandFn
	t.Cleanup(func() {
		sendUIControlCommandFn = prevSendUIControlCommandFn
	})

	var commands []string
	sendUIControlCommandFn = func(command string) error {
		commands = append(commands, command)
		return nil
	}

	controller := newHelperController(nil)
	controller.hideMainWindow()

	require.Equal(t, []string{controlCommandQuit}, commands)
}

func TestHelperControllerStopUIClearsTrackedProcessAfterQuit(t *testing.T) {
	prevSendUIControlCommandFn := sendUIControlCommandFn
	t.Cleanup(func() {
		sendUIControlCommandFn = prevSendUIControlCommandFn
	})

	sendUIControlCommandFn = func(command string) error {
		require.Equal(t, controlCommandQuit, command)
		return nil
	}

	controller := newHelperController(nil)
	controller.uiCmd = &exec.Cmd{}

	require.NoError(t, controller.stopUI())

	controller.uiMu.Lock()
	defer controller.uiMu.Unlock()
	assert.Nil(t, controller.uiCmd)
}
