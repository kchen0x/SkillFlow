package adapters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/memorycatalog/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildExplicitRulesIndexUsesMarkdownRelativePaths(t *testing.T) {
	modules := []*domain.ModuleMemory{
		{Name: "testing"},
		{Name: "style"},
	}
	memoryPath := filepath.Join("workspace", "agent", "AGENTS.md")
	rulesDir := filepath.Join("workspace", "agent", "rules")

	index := buildExplicitRulesIndex(modules, memoryPath, rulesDir)

	require.Equal(t, []string{
		"[testing](rules/sf-testing.md)",
		"[style](rules/sf-style.md)",
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
