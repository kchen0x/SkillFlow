package upgrade_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/upgrade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunMigratesToolTerminologyInConfigFiles(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{
  "defaultCategory": "Default",
  "logLevel": "info",
  "repoScanMaxDepth": 5,
  "skillStatusVisibility": {
    "mySkills": ["updatable", "pushedTools"],
    "myTools": ["imported", "updatable", "pushedTools"],
    "pushToTool": ["pushedTools"],
    "pullFromTool": ["imported"],
    "starredRepos": ["imported", "pushedTools"],
    "githubInstall": ["imported", "updatable", "pushedTools"]
  },
  "tools": [
    { "name": "codex", "enabled": true }
  ]
}`), 0o644))

	require.NoError(t, os.WriteFile(filepath.Join(dir, "config_local.json"), []byte(`{
  "skillsStorageDir": "/tmp/skills",
  "autoPushTools": ["codex", "gemini-cli"],
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

	require.NoError(t, upgrade.Run(dir))

	shared := readJSONMap(t, filepath.Join(dir, "config.json"))
	require.Contains(t, shared, "agents")
	require.NotContains(t, shared, "tools")

	sharedAgents := requireJSONArrayMaps(t, shared["agents"])
	require.Len(t, sharedAgents, 1)
	assert.Equal(t, "codex", sharedAgents[0]["name"])
	assert.Equal(t, true, sharedAgents[0]["enabled"])

	visibility := requireJSONMap(t, shared["skillStatusVisibility"])
	require.Contains(t, visibility, "myAgents")
	require.Contains(t, visibility, "pushToAgent")
	require.Contains(t, visibility, "pullFromAgent")
	require.NotContains(t, visibility, "githubInstall")
	require.NotContains(t, visibility, "myTools")
	require.NotContains(t, visibility, "pushToTool")
	require.NotContains(t, visibility, "pullFromTool")
	assert.Equal(t, []any{"updatable", "pushedAgents"}, requireJSONArray(t, visibility["mySkills"]))
	assert.Equal(t, []any{"imported", "updatable", "pushedAgents"}, requireJSONArray(t, visibility["myAgents"]))
	assert.Equal(t, []any{"pushedAgents"}, requireJSONArray(t, visibility["pushToAgent"]))
	assert.Equal(t, []any{"imported"}, requireJSONArray(t, visibility["pullFromAgent"]))

	local := readJSONMap(t, filepath.Join(dir, "config_local.json"))
	require.Contains(t, local, "agents")
	require.NotContains(t, local, "tools")
	require.Contains(t, local, "autoPushAgents")
	require.NotContains(t, local, "autoPushTools")
	assert.Equal(t, []any{"codex", "gemini-cli"}, requireJSONArray(t, local["autoPushAgents"]))

	localAgents := requireJSONArrayMaps(t, local["agents"])
	require.Len(t, localAgents, 1)
	assert.Equal(t, "codex", localAgents[0]["name"])
	assert.Equal(t, []any{"/tmp/codex/skills"}, requireJSONArray(t, localAgents[0]["scanDirs"]))
	assert.Equal(t, "/tmp/codex/skills", localAgents[0]["pushDir"])
}

func TestRunIsNoOpForAlreadyUpgradedConfigFiles(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{
  "defaultCategory": "Default",
  "skillStatusVisibility": {
    "mySkills": ["updatable", "pushedAgents"],
    "myAgents": ["imported", "updatable", "pushedAgents"],
    "pushToAgent": ["pushedAgents"],
    "pullFromAgent": ["imported"],
    "starredRepos": ["imported", "pushedAgents"]
  },
  "agents": [
    { "name": "codex", "enabled": true }
  ]
}`), 0o644))

	require.NoError(t, os.WriteFile(filepath.Join(dir, "config_local.json"), []byte(`{
  "skillsStorageDir": "/tmp/skills",
  "autoPushAgents": ["codex"],
  "agents": [
    {
      "name": "codex",
      "scanDirs": ["/tmp/codex/skills"],
      "pushDir": "/tmp/codex/skills",
      "custom": false,
      "enabled": true
    }
  ]
}`), 0o644))

	beforeShared, err := os.ReadFile(filepath.Join(dir, "config.json"))
	require.NoError(t, err)
	beforeLocal, err := os.ReadFile(filepath.Join(dir, "config_local.json"))
	require.NoError(t, err)

	require.NoError(t, upgrade.Run(dir))
	require.NoError(t, upgrade.Run(dir))

	afterShared, err := os.ReadFile(filepath.Join(dir, "config.json"))
	require.NoError(t, err)
	afterLocal, err := os.ReadFile(filepath.Join(dir, "config_local.json"))
	require.NoError(t, err)

	assert.Equal(t, string(beforeShared), string(afterShared))
	assert.Equal(t, string(beforeLocal), string(afterLocal))
}

func TestRunRemovesGitHubInstallVisibilityFromExistingSharedConfig(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{
  "defaultCategory": "Default",
  "skillStatusVisibility": {
    "mySkills": ["updatable", "pushedAgents"],
    "myAgents": ["imported", "updatable", "pushedAgents"],
    "pushToAgent": ["pushedAgents"],
    "pullFromAgent": ["imported"],
    "starredRepos": ["imported", "pushedAgents"],
    "githubInstall": ["imported", "updatable", "pushedAgents"]
  }
}`), 0o644))

	require.NoError(t, upgrade.Run(dir))

	shared := readJSONMap(t, filepath.Join(dir, "config.json"))
	visibility := requireJSONMap(t, shared["skillStatusVisibility"])
	require.NotContains(t, visibility, "githubInstall")
}

func TestRunMigratesLegacyToolsWhenAgentsArrayAlreadyExistsButIsEmpty(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{
  "defaultCategory": "Default",
  "agents": [],
  "tools": [
    { "name": "codex", "enabled": true }
  ]
}`), 0o644))

	require.NoError(t, os.WriteFile(filepath.Join(dir, "config_local.json"), []byte(`{
  "skillsStorageDir": "/tmp/skills",
  "agents": [],
  "tools": [
    {
      "name": "codex",
      "scanDirs": ["/tmp/codex/skills"],
      "pushDir": "/tmp/codex/skills",
      "custom": false,
      "enabled": true
    }
  ],
  "autoPushAgents": [],
  "autoPushTools": ["codex"]
}`), 0o644))

	require.NoError(t, upgrade.Run(dir))

	shared := readJSONMap(t, filepath.Join(dir, "config.json"))
	sharedAgents := requireJSONArrayMaps(t, shared["agents"])
	require.Len(t, sharedAgents, 1)
	assert.Equal(t, "codex", sharedAgents[0]["name"])
	assert.Equal(t, true, sharedAgents[0]["enabled"])
	require.NotContains(t, shared, "tools")

	local := readJSONMap(t, filepath.Join(dir, "config_local.json"))
	localAgents := requireJSONArrayMaps(t, local["agents"])
	require.Len(t, localAgents, 1)
	assert.Equal(t, "codex", localAgents[0]["name"])
	assert.Equal(t, []any{"/tmp/codex/skills"}, requireJSONArray(t, localAgents[0]["scanDirs"]))
	assert.Equal(t, "/tmp/codex/skills", localAgents[0]["pushDir"])
	assert.Equal(t, []any{"codex"}, requireJSONArray(t, local["autoPushAgents"]))
	require.NotContains(t, local, "tools")
	require.NotContains(t, local, "autoPushTools")
}

func TestRunFailsOnMalformedConfigJSON(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"tools": [`), 0o644))

	err := upgrade.Run(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config.json")
}

func readJSONMap(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var out map[string]any
	require.NoError(t, json.Unmarshal(data, &out))
	return out
}

func requireJSONMap(t *testing.T, value any) map[string]any {
	t.Helper()
	out, ok := value.(map[string]any)
	require.True(t, ok)
	return out
}

func requireJSONArray(t *testing.T, value any) []any {
	t.Helper()
	out, ok := value.([]any)
	require.True(t, ok)
	return out
}

func requireJSONArrayMaps(t *testing.T, value any) []map[string]any {
	t.Helper()
	items := requireJSONArray(t, value)
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		out = append(out, requireJSONMap(t, item))
	}
	return out
}
