package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shinerio/skillflow/core/config"
	"github.com/shinerio/skillflow/core/platform/upgrade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartupBackgroundTaskPlanUsesStaggeredDelays(t *testing.T) {
	app := NewApp()

	tasks := app.startupBackgroundTaskPlan()

	require.Len(t, tasks, 4)
	assert.Equal(t, "git.pull", tasks[0].name)
	assert.Equal(t, 750*time.Millisecond, tasks[0].delay)
	assert.Equal(t, "starred.refresh", tasks[1].name)
	assert.Equal(t, 3*time.Second, tasks[1].delay)
	assert.Equal(t, "skills.check_updates", tasks[2].name)
	assert.Equal(t, 5250*time.Millisecond, tasks[2].delay)
	assert.Equal(t, "app.check_update", tasks[3].name)
	assert.Equal(t, 8*time.Second, tasks[3].delay)
}

func TestScheduleStartupBackgroundTasksRegistersAllTasks(t *testing.T) {
	scheduled := make([]time.Duration, 0, 2)
	executed := make([]string, 0, 2)

	scheduleStartupBackgroundTasks([]startupBackgroundTask{
		{name: "first", delay: 250 * time.Millisecond, run: func() { executed = append(executed, "first") }},
		{name: "second", delay: time.Second, run: func() { executed = append(executed, "second") }},
	}, func(task startupBackgroundTask) {
		scheduled = append(scheduled, task.delay)
		task.run()
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
