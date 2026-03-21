package upgrade_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shinerio/skillflow/core/platform/upgrade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunRemovesLegacySkillsStorageDirAndBackfillsRepoCacheDir(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "config_local.json"), []byte(`{
  "skillsStorageDir": "/tmp/skills",
  "agents": []
}`), 0o644))

	require.NoError(t, upgrade.Run(dir))

	localData, err := os.ReadFile(filepath.Join(dir, "config_local.json"))
	require.NoError(t, err)
	assert.Contains(t, string(localData), `"repoCacheDir": "`+filepath.ToSlash(filepath.Join(dir, "cache", "repos"))+`"`)
	assert.NotContains(t, string(localData), "skillsStorageDir")
}

func TestRunRemovesLegacyStarRepoLocalDir(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "star_repos.json"), []byte(`[
  {
    "url": "https://github.com/example/demo",
    "name": "example/demo",
    "source": "github.com/example/demo",
    "localDir": "cache/repos/github.com/example/demo"
  }
]`), 0o644))

	require.NoError(t, upgrade.Run(dir))

	starData, err := os.ReadFile(filepath.Join(dir, "star_repos.json"))
	require.NoError(t, err)
	assert.False(t, strings.Contains(string(starData), "localDir"), "unexpected localDir in upgraded star_repos.json: %s", string(starData))
}
