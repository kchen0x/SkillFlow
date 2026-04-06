package daemon

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shinerio/skillflow/core/config"
	memorycatalogapp "github.com/shinerio/skillflow/core/memorycatalog/app"
	"github.com/shinerio/skillflow/core/platform/upgrade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRuntimeRunsUpgradeBeforeConfigLoad(t *testing.T) {
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

	callOrder := make([]string, 0, 2)
	rt, err := NewRuntime(dir, Dependencies{
		RunUpgrade: func(dataDir string) error {
			callOrder = append(callOrder, "upgrade")
			return upgrade.Run(dataDir)
		},
		LoadConfig: func(dataDir string) (*config.Service, config.AppConfig, error) {
			callOrder = append(callOrder, "load")
			svc := config.NewService(dataDir)
			cfg, err := svc.Load()
			return svc, cfg, err
		},
		NewMemoryServices: func(_ *config.Service, _ string) (*memorycatalogapp.MemoryService, *memorycatalogapp.PushService) {
			return &memorycatalogapp.MemoryService{}, &memorycatalogapp.PushService{}
		},
		RegisterAdapters:  func() {},
		RegisterProviders: func() {},
	})
	require.NoError(t, err)

	assert.Equal(t, []string{"upgrade", "load"}, callOrder)
	assert.Equal(t, "Default", rt.ConfigSnapshot.DefaultCategory)
	assert.Contains(t, rt.ConfigSnapshot.AutoPushAgents, "codex")
}

func TestNewRuntimeInitializesServicesWithoutWailsContext(t *testing.T) {
	dir := t.TempDir()
	cfgService := config.NewService(dir)
	cfg := config.DefaultConfig(dir)
	cfg.RepoCacheDir = filepath.Join(dir, "cache", "repos")

	registerAdaptersCalls := 0
	registerProvidersCalls := 0
	memoryService := &memorycatalogapp.MemoryService{}
	pushService := &memorycatalogapp.PushService{}

	rt, err := NewRuntime(dir, Dependencies{
		LoadConfig: func(dataDir string) (*config.Service, config.AppConfig, error) {
			return cfgService, cfg, nil
		},
		SyncLaunchAtLogin: func(enabled bool) error {
			assert.False(t, enabled)
			return nil
		},
		NewMemoryServices: func(gotCfg *config.Service, gotDir string) (*memorycatalogapp.MemoryService, *memorycatalogapp.PushService) {
			require.Same(t, cfgService, gotCfg)
			require.Equal(t, dir, gotDir)
			return memoryService, pushService
		},
		RegisterAdapters: func() {
			registerAdaptersCalls++
		},
		RegisterProviders: func() {
			registerProvidersCalls++
		},
	})
	require.NoError(t, err)

	require.NotNil(t, rt)
	require.NotNil(t, rt.Hub)
	require.NotNil(t, rt.ConfigService)
	require.NotNil(t, rt.Storage)
	require.NotNil(t, rt.StarStorage)
	require.NotNil(t, rt.ViewCache)
	assert.Equal(t, filepath.Join(dir, "cache"), rt.CacheDir)
	assert.Same(t, memoryService, rt.MemoryService)
	assert.Same(t, pushService, rt.MemoryPushService)
	assert.Equal(t, 1, registerAdaptersCalls)
	assert.Equal(t, 1, registerProvidersCalls)
	assert.Nil(t, rt.ConfigLoadErr)
}

func TestRuntimeScheduleStartupTasksRunsOnlyOnce(t *testing.T) {
	rt := &Runtime{}
	calls := 0
	tasks := []StartupTask{
		{Name: "first"},
		{Name: "second"},
	}

	rt.ScheduleStartupTasks(tasks, func(task StartupTask) {
		calls++
	})
	rt.ScheduleStartupTasks(tasks, func(task StartupTask) {
		calls++
	})

	assert.Equal(t, 2, calls)
}

func TestRuntimeStartAutoSyncTimerDisablesExistingTimerWhenIntervalIsZero(t *testing.T) {
	stopped := make(chan struct{})
	rt := &Runtime{
		stopAutoSync: stopped,
	}

	rt.StartAutoSyncTimer(0, func() {})

	select {
	case <-stopped:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected existing auto sync timer to stop when interval is disabled")
	}
	assert.Nil(t, rt.stopAutoSync)
}
