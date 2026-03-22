package config_test

import (
	"os"
	"path/filepath"
	"testing"

	agentapp "github.com/shinerio/skillflow/core/agentintegration/app"
	"github.com/shinerio/skillflow/core/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfigUsesAgentTerminology(t *testing.T) {
	dir := t.TempDir()

	cfg := config.DefaultConfig(dir)

	assert.NotEmpty(t, cfg.Agents)
}

func TestSaveAndLoadConfigWithAgentTerminology(t *testing.T) {
	dir := t.TempDir()
	svc := config.NewService(dir)
	cfg := config.DefaultConfig(dir)
	cfg.AutoPushAgents = []string{"codex", "gemini-cli"}

	require.NoError(t, svc.Save(cfg))

	loaded, err := svc.Load()
	require.NoError(t, err)
	assert.Equal(t, []string{"codex", "gemini-cli"}, loaded.AutoPushAgents)

	sharedData, err := os.ReadFile(filepath.Join(dir, "config.json"))
	require.NoError(t, err)
	assert.Contains(t, string(sharedData), `"agents"`)
	assert.NotContains(t, string(sharedData), `"skillStatusVisibility"`)
	assert.NotContains(t, string(sharedData), `"tools"`)

	localData, err := os.ReadFile(filepath.Join(dir, "config_local.json"))
	require.NoError(t, err)
	assert.Contains(t, string(localData), `"agents"`)
	assert.Contains(t, string(localData), `"autoPushAgents"`)
	assert.NotContains(t, string(localData), `"tools"`)
	assert.NotContains(t, string(localData), `"autoPushTools"`)
}

func TestConfigAgentsCanBeConsumedDirectlyByAgentIntegration(t *testing.T) {
	dir := t.TempDir()

	cfg := config.DefaultConfig(dir)

	profile, ok := agentapp.FindProfile(cfg.Agents, "codex")
	require.True(t, ok)
	assert.Equal(t, "codex", profile.Name)
	assert.True(t, profile.Enabled)
}
