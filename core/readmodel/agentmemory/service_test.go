package agentmemory

import (
	"os"
	"path/filepath"
	"testing"

	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadPreviewReadsMainMemoryAndRules(t *testing.T) {
	dir := t.TempDir()
	memoryPath := filepath.Join(dir, "AGENTS.md")
	rulesDir := filepath.Join(dir, "rules")

	require.NoError(t, os.MkdirAll(rulesDir, 0o755))
	require.NoError(t, os.WriteFile(memoryPath, []byte("# Agent memory\nUse Go.\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(rulesDir, "alpha.md"), []byte("alpha"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(rulesDir, "sf-beta.md"), []byte("beta"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(rulesDir, "notes.txt"), []byte("skip"), 0o644))

	preview, err := LoadPreview(agentdomain.AgentProfile{
		Name:       "codex",
		MemoryPath: memoryPath,
		RulesDir:   rulesDir,
	})
	require.NoError(t, err)

	assert.Equal(t, "codex", preview.AgentName)
	assert.Equal(t, memoryPath, preview.MemoryPath)
	assert.Equal(t, rulesDir, preview.RulesDir)
	assert.True(t, preview.MainExists)
	assert.True(t, preview.RulesDirExists)
	assert.Equal(t, "# Agent memory\nUse Go.\n", preview.MainContent)
	require.Len(t, preview.Rules, 2)
	assert.Equal(t, "alpha.md", preview.Rules[0].Name)
	assert.Equal(t, "sf-beta.md", preview.Rules[1].Name)
	assert.Equal(t, "alpha", preview.Rules[0].Content)
	assert.Equal(t, "beta", preview.Rules[1].Content)
}

func TestLoadPreviewMarksManagedRuleFiles(t *testing.T) {
	dir := t.TempDir()
	rulesDir := filepath.Join(dir, "rules")

	require.NoError(t, os.MkdirAll(rulesDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(rulesDir, "sf-managed.md"), []byte("managed"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(rulesDir, "custom.md"), []byte("custom"), 0o644))

	preview, err := LoadPreview(agentdomain.AgentProfile{
		Name:     "codex",
		RulesDir: rulesDir,
	})
	require.NoError(t, err)
	require.Len(t, preview.Rules, 2)

	assert.False(t, preview.Rules[0].Managed)
	assert.Equal(t, "custom.md", preview.Rules[0].Name)
	assert.True(t, preview.Rules[1].Managed)
	assert.Equal(t, "sf-managed.md", preview.Rules[1].Name)
}
