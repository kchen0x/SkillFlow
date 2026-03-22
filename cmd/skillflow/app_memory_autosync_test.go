package main

import (
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/config"
	memorydomain "github.com/shinerio/skillflow/core/memorycatalog/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncMemoryToAutoPushAgentsPushesAllModulesToEnabledAgents(t *testing.T) {
	dataDir := t.TempDir()
	svc := config.NewService(dataDir)
	cfg := config.DefaultConfig(dataDir)
	cfg.Agents = []config.AgentConfig{
		{
			Name:       "codex",
			MemoryPath: filepath.Join(dataDir, "codex", "AGENTS.md"),
			RulesDir:   filepath.Join(dataDir, "codex", "rules"),
			Enabled:    true,
		},
	}
	require.NoError(t, svc.Save(cfg))

	app := NewApp()
	app.config = svc
	app.memoryService, app.memoryPushService = newMemoryServices(app)

	_, err := app.memoryService.SaveMainMemory("Main memory")
	require.NoError(t, err)
	_, err = app.memoryService.CreateModule("style", "Style rules")
	require.NoError(t, err)
	_, err = app.memoryService.CreateModule("testing", "Always test")
	require.NoError(t, err)
	require.NoError(t, app.memoryService.SavePushConfig(memorydomain.MemoryPushConfig{
		AgentType: "codex",
		Mode:      memorydomain.PushModeMerge,
		AutoPush:  true,
	}))

	require.NoError(t, app.syncMemoryToAutoPushAgents())

	assert.FileExists(t, filepath.Join(dataDir, "codex", "AGENTS.md"))
	assert.FileExists(t, filepath.Join(dataDir, "codex", "rules", "sf-style.md"))
	assert.FileExists(t, filepath.Join(dataDir, "codex", "rules", "sf-testing.md"))
}

func TestSaveMemoryPushConfigAutoPushesExistingMemoryImmediately(t *testing.T) {
	dataDir := t.TempDir()
	svc := config.NewService(dataDir)
	cfg := config.DefaultConfig(dataDir)
	cfg.Agents = []config.AgentConfig{
		{
			Name:       "codex",
			MemoryPath: filepath.Join(dataDir, "codex", "AGENTS.md"),
			RulesDir:   filepath.Join(dataDir, "codex", "rules"),
			Enabled:    true,
		},
	}
	require.NoError(t, svc.Save(cfg))

	app := NewApp()
	app.config = svc
	app.memoryService, app.memoryPushService = newMemoryServices(app)

	_, err := app.memoryService.SaveMainMemory("Main memory")
	require.NoError(t, err)
	_, err = app.memoryService.CreateModule("style", "Style rules")
	require.NoError(t, err)

	require.NoError(t, app.SaveMemoryPushConfig("codex", string(memorydomain.PushModeMerge), true))

	assert.FileExists(t, filepath.Join(dataDir, "codex", "AGENTS.md"))
	assert.FileExists(t, filepath.Join(dataDir, "codex", "rules", "sf-style.md"))
}
