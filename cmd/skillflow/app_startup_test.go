package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shinerio/skillflow/core/config"
	daemonruntime "github.com/shinerio/skillflow/core/platform/daemon"
	"github.com/shinerio/skillflow/core/platform/eventbus"
	"github.com/shinerio/skillflow/core/platform/upgrade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartupBackgroundTaskPlanUsesStaggeredDelays(t *testing.T) {
	app := NewApp()

	tasks := app.startupBackgroundTaskPlan()

	require.Len(t, tasks, 4)
	assert.Equal(t, "git.pull", tasks[0].Name)
	assert.Equal(t, 750*time.Millisecond, tasks[0].Delay)
	assert.Equal(t, "starred.refresh", tasks[1].Name)
	assert.Equal(t, 3*time.Second, tasks[1].Delay)
	assert.Equal(t, "skills.check_updates", tasks[2].Name)
	assert.Equal(t, 5250*time.Millisecond, tasks[2].Delay)
	assert.Equal(t, "app.check_update", tasks[3].Name)
	assert.Equal(t, 8*time.Second, tasks[3].Delay)
}

func TestScheduleStartupBackgroundTasksRegistersAllTasks(t *testing.T) {
	scheduled := make([]time.Duration, 0, 2)
	executed := make([]string, 0, 2)

	scheduleStartupBackgroundTasks([]startupBackgroundTask{
		{Name: "first", Delay: 250 * time.Millisecond, Run: func() { executed = append(executed, "first") }},
		{Name: "second", Delay: time.Second, Run: func() { executed = append(executed, "second") }},
	}, func(task startupBackgroundTask) {
		scheduled = append(scheduled, task.Delay)
		task.Run()
	})

	assert.Equal(t, []time.Duration{250 * time.Millisecond, time.Second}, scheduled)
	assert.Equal(t, []string{"first", "second"}, executed)
}

func TestStartupRunsUpgradeBeforeConfigLoad(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{
  "defaultCategory": "Default",
  "skillStatusVisibility": {
    "mySkills": ["updatable", "pushedTools"],
    "myTools": ["imported", "updatable", "pushedTools"],
    "pushToTool": ["pushedTools"],
    "pullFromTool": ["imported"],
    "starredRepos": ["imported", "pushedTools"]
  },
  "tools": [
    { "name": "codex", "enabled": true }
  ]
}`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config_local.json"), []byte(`{
  "skillsStorageDir": "/tmp/skills",
  "autoPushTools": ["codex"],
  "tools": [
    {
      "name": "codex",
      "scanDirs": ["/tmp/codex/skills"],
      "pushDir": "/tmp/codex/skills",
      "custom": false,
      "enabled": true
    }
  ]
}`), 0o644))

	prevAppDataDirFunc := appDataDirFunc
	prevRunStartupUpgrade := runStartupUpgrade
	prevLoadStartupConfig := loadStartupConfig
	t.Cleanup(func() {
		appDataDirFunc = prevAppDataDirFunc
		runStartupUpgrade = prevRunStartupUpgrade
		loadStartupConfig = prevLoadStartupConfig
	})

	callOrder := make([]string, 0, 2)
	appDataDirFunc = func() string {
		return dir
	}
	runStartupUpgrade = func(dataDir string) error {
		callOrder = append(callOrder, "upgrade")
		return upgrade.Run(dataDir)
	}
	loadStartupConfig = func(dataDir string) (*config.Service, config.AppConfig, error) {
		callOrder = append(callOrder, "load")
		svc := config.NewService(dataDir)
		cfg, err := svc.Load()
		return svc, cfg, err
	}

	controller := &fakeLaunchAtLoginController{}
	app := NewApp()
	app.autostartFactory = func() (launchAtLoginController, error) {
		return controller, nil
	}

	app.startup(nil)

	assert.Equal(t, []string{"upgrade", "load"}, callOrder)

	sharedData, err := os.ReadFile(filepath.Join(dir, "config.json"))
	require.NoError(t, err)
	assert.Contains(t, string(sharedData), `"agents"`)
	assert.NotContains(t, string(sharedData), `"tools"`)
	assert.NotContains(t, string(sharedData), `"githubInstall"`)

	localData, err := os.ReadFile(filepath.Join(dir, "config_local.json"))
	require.NoError(t, err)
	assert.Contains(t, string(localData), `"agents"`)
	assert.Contains(t, string(localData), `"autoPushAgents"`)
	assert.Contains(t, string(localData), `"repoCacheDir"`)
	assert.NotContains(t, string(localData), `"autoPushTools"`)
	assert.NotContains(t, string(localData), `"skillsStorageDir"`)
}

func TestStartupUsesDaemonRuntimeBuilder(t *testing.T) {
	dir := t.TempDir()

	prevAppDataDirFunc := appDataDirFunc
	prevNewDaemonRuntimeFn := newDaemonRuntimeFn
	t.Cleanup(func() {
		appDataDirFunc = prevAppDataDirFunc
		newDaemonRuntimeFn = prevNewDaemonRuntimeFn
	})

	appDataDirFunc = func() string {
		return dir
	}

	expectedHub := eventbus.NewHub()
	expectedConfig := config.NewService(dir)
	expectedRuntime := &daemonruntime.Runtime{
		DataDir:       dir,
		ConfigService: expectedConfig,
		ConfigSnapshot: config.AppConfig{
			Cloud: config.CloudConfig{},
		},
		Hub: expectedHub,
	}

	called := 0
	newDaemonRuntimeFn = func(dataDir string, deps daemonruntime.Dependencies) (*daemonruntime.Runtime, error) {
		called++
		assert.Equal(t, dir, dataDir)
		require.NotNil(t, deps.SyncLaunchAtLogin)
		return expectedRuntime, nil
	}

	app := NewApp()
	app.autostartFactory = func() (launchAtLoginController, error) {
		t.Fatal("startup should delegate launch-at-login reconciliation to daemon runtime builder")
		return nil, nil
	}

	app.startup(nil)

	assert.Equal(t, 1, called)
	assert.Same(t, expectedRuntime, app.backendRuntime)
	assert.Same(t, expectedHub, app.hub)
	assert.Same(t, expectedConfig, app.config)
}

func TestStartupSkipsAutoSyncTimerForUIProcessRole(t *testing.T) {
	dir := t.TempDir()

	prevAppDataDirFunc := appDataDirFunc
	prevNewDaemonRuntimeFn := newDaemonRuntimeFn
	prevStartAppAutoSyncTimerFn := startAppAutoSyncTimerFn
	prevActiveProcessRole := activeProcessRole
	t.Cleanup(func() {
		appDataDirFunc = prevAppDataDirFunc
		newDaemonRuntimeFn = prevNewDaemonRuntimeFn
		startAppAutoSyncTimerFn = prevStartAppAutoSyncTimerFn
		activeProcessRole = prevActiveProcessRole
	})

	appDataDirFunc = func() string { return dir }
	activeProcessRole = processRoleUI

	svc := config.NewService(dir)
	cfg := config.DefaultConfig(dir)
	cfg.LogLevel = config.LogLevelDebug
	require.NoError(t, svc.Save(cfg))

	daemonRuntimeCalls := 0
	newDaemonRuntimeFn = func(dataDir string, deps daemonruntime.Dependencies) (*daemonruntime.Runtime, error) {
		daemonRuntimeCalls++
		return &daemonruntime.Runtime{
			DataDir:         dataDir,
			ConfigService:   config.NewService(dataDir),
			ConfigSnapshot:  config.AppConfig{Cloud: config.CloudConfig{SyncIntervalMinutes: 13}},
			Hub:             eventbus.NewHub(),
		}, nil
	}

	autoSyncCalls := 0
	startAppAutoSyncTimerFn = func(app *App, minutes int) {
		autoSyncCalls++
	}

	app := NewApp()
	app.startup(nil)

	assert.Zero(t, autoSyncCalls)
	assert.Zero(t, daemonRuntimeCalls)
	assert.Nil(t, app.backendRuntime)
	require.NotNil(t, app.config)
	assert.Equal(t, dir, app.config.DataDir())
	assert.NotNil(t, app.sysLog)
	assert.NotNil(t, app.storage)
	assert.NotNil(t, app.starStorage)
}
