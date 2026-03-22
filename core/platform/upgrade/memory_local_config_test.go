package upgrade_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/platform/upgrade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunRemovesLegacyMemoryModuleTargetsFromLocalConfig(t *testing.T) {
	dir := t.TempDir()
	memoryDir := filepath.Join(dir, "memory")
	require.NoError(t, os.MkdirAll(memoryDir, 0o755))

	payload := map[string]any{
		"pushConfigs": map[string]any{
			"codex": map[string]any{
				"mode":     "merge",
				"autoPush": true,
			},
		},
		"modules": map[string]any{
			"testing-rules": map[string]any{
				"pushTargets": []string{"codex"},
			},
		},
		"pushState": map[string]any{
			"codex": map[string]any{
				"lastPushedHash": "abc",
			},
		},
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(memoryDir, "memory_local.json"), data, 0o644))

	require.NoError(t, upgrade.Run(dir))

	upgraded, err := os.ReadFile(filepath.Join(memoryDir, "memory_local.json"))
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(upgraded, &decoded))

	_, hasModules := decoded["modules"]
	assert.False(t, hasModules, "legacy modules config should be removed after upgrade")
	assert.Contains(t, decoded, "pushConfigs")
	assert.Contains(t, decoded, "pushState")
}
