package domain_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/agentintegration/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuiltinAgentNamesAreStable(t *testing.T) {
	assert.Equal(t, []string{"claude-code", "opencode", "codex", "gemini-cli", "openclaw"}, domain.BuiltinAgentNames())
}

func TestDefaultProfileUsesExpectedDirectories(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	claude := domain.DefaultProfile("claude-code")
	require.True(t, claude.Enabled)
	assert.False(t, claude.Custom)
	assert.Equal(t, filepath.Join(home, ".claude", "skills"), claude.PushDir)
	assert.Equal(t, []string{
		filepath.Join(home, ".claude", "skills"),
		filepath.Join(home, ".claude", "plugins", "marketplaces"),
	}, claude.ScanDirs)

	codex := domain.DefaultProfile("codex")
	assert.Equal(t, filepath.Join(home, ".agents", "skills"), codex.PushDir)
	assert.Equal(t, []string{filepath.Join(home, ".agents", "skills")}, codex.ScanDirs)
}

func TestDefaultProfileReturnsEmptyForUnknownAgent(t *testing.T) {
	profile := domain.DefaultProfile("unknown")

	assert.Equal(t, "unknown", profile.Name)
	assert.Empty(t, profile.ScanDirs)
	assert.Empty(t, profile.PushDir)
	assert.True(t, profile.Enabled)
	assert.False(t, profile.Custom)
}
