package adapters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/memorycatalog/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildExplicitRulesIndexUsesAbsolutePaths(t *testing.T) {
	modules := []*domain.ModuleMemory{
		{Name: "testing"},
		{Name: "style"},
	}
	rulesDir := filepath.Join("workspace", "agent", "rules")

	index := buildExplicitRulesIndex(modules, rulesDir)

	require.Equal(t, []string{
		"[testing](" + filepath.ToSlash(filepath.Join(rulesDir, "sf-testing.md")) + ")",
		"[style](" + filepath.ToSlash(filepath.Join(rulesDir, "sf-style.md")) + ")",
	}, index.Entries)
}

func TestPushMainMemoryMergeUsesTaggedManagedSections(t *testing.T) {
	tmpDir := t.TempDir()
	memoryPath := filepath.Join(tmpDir, "AGENTS.md")
	adapter := &baseAdapter{}
	content := "<skillflow-managed>\nmain\n</skillflow-managed>\n\n<skillflow-module>\n[module](rules/sf-module.md)\n</skillflow-module>"

	require.NoError(t, adapter.PushMainMemory(content, domain.PushModeMerge, memoryPath))

	data, err := os.ReadFile(memoryPath)
	require.NoError(t, err)
	assert.Equal(t, content, string(data))
}
