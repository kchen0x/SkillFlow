package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultSettingsIncludeBuiltinAgents(t *testing.T) {
	settings := DefaultSettings()
	assert.Equal(t, DefaultRepoScanMaxDepth, settings.Shared.RepoScanMaxDepth)
	assert.NotEmpty(t, settings.Shared.Agents)
	assert.NotEmpty(t, settings.Local.Agents)
}

func TestNormalizeAutoPushAgentNames(t *testing.T) {
	normalized := NormalizeAutoPushAgentNames([]string{" codex ", "gemini-cli", "codex", ""})
	assert.Equal(t, []string{"codex", "gemini-cli"}, normalized)
}

func TestNormalizeRepoScanMaxDepth(t *testing.T) {
	assert.Equal(t, DefaultRepoScanMaxDepth, NormalizeRepoScanMaxDepth(0))
	assert.Equal(t, MaxRepoScanMaxDepth, NormalizeRepoScanMaxDepth(999))
	assert.Equal(t, 7, NormalizeRepoScanMaxDepth(7))
}
