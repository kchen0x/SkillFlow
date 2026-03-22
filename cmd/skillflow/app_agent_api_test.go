package main

import (
	"os"
	"path/filepath"
	"testing"

	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
	"github.com/shinerio/skillflow/core/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEnabledAgentsReturnsEnabledAgents(t *testing.T) {
	app, _, _ := newAutoPushTestApp(t, []string{"codex"})

	agents, err := app.GetEnabledAgents()
	require.NoError(t, err)
	require.NotEmpty(t, agents)

	enabledNames := make([]string, 0, len(agents))
	for _, agent := range agents {
		assert.True(t, agent.Enabled)
		enabledNames = append(enabledNames, agent.Name)
	}
	assert.Contains(t, enabledNames, "codex")
}

func TestPushConflictUsesAgentNameField(t *testing.T) {
	conflict := agentdomain.PushConflict{
		SkillName:  "demo-skill",
		AgentName:  "codex",
		TargetPath: "/tmp/codex/demo-skill",
	}

	assert.Equal(t, "codex", conflict.AgentName)
}

func TestGetAgentMemoryPreviewReturnsAgentPreview(t *testing.T) {
	dataDir := t.TempDir()
	memoryPath := filepath.Join(dataDir, "codex", "AGENTS.md")
	rulesDir := filepath.Join(dataDir, "codex", "rules")
	require.NoError(t, os.MkdirAll(rulesDir, 0o755))
	require.NoError(t, os.WriteFile(memoryPath, []byte("main memory"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(rulesDir, "sf-style.md"), []byte("style"), 0o644))

	svc := config.NewService(dataDir)
	cfg := config.DefaultConfig(dataDir)
	cfg.Agents = []config.AgentConfig{{
		Name:       "codex",
		ScanDirs:   []string{filepath.Join(dataDir, "skills")},
		PushDir:    filepath.Join(dataDir, "skills"),
		MemoryPath: memoryPath,
		RulesDir:   rulesDir,
		Enabled:    true,
	}}
	require.NoError(t, svc.Save(cfg))

	app := NewApp()
	app.config = svc

	preview, err := app.GetAgentMemoryPreview("codex")
	require.NoError(t, err)
	require.NotNil(t, preview)
	assert.Equal(t, "codex", preview.AgentName)
	assert.Equal(t, memoryPath, preview.MemoryPath)
	assert.Equal(t, rulesDir, preview.RulesDir)
	assert.True(t, preview.MainExists)
	assert.Equal(t, "main memory", preview.MainContent)
	require.Len(t, preview.Rules, 1)
	assert.Equal(t, "sf-style.md", preview.Rules[0].Name)
	assert.True(t, preview.Rules[0].Managed)
}
