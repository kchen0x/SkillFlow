package main

import (
	"context"
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
	prevStartAppAutoSyncTimerFn := startAppAutoSyncTimerFn
	prevStartAppBackgroundTasksFn := startAppBackgroundTasksFn
	prevActiveProcessRole := activeProcessRole
	t.Cleanup(func() {
		newDaemonRuntimeFn = prevNewDaemonRuntimeFn
		newDaemonAppFn = prevNewDaemonAppFn
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

	autoSyncCalls := 0
	backgroundCalls := 0
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
	assert.Equal(t, 1, autoSyncCalls)
	assert.Equal(t, 1, backgroundCalls)
}
