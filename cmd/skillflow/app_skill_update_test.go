package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/config"
	coregit "github.com/shinerio/skillflow/core/git"
	"github.com/shinerio/skillflow/core/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckUpdatesUsesLocalCacheSHA(t *testing.T) {
	app, _, _, dataDir := newUpdateSkillTestApp(t)
	sourceDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nOld\n")
	_, oldSHA, newSHA := seedCachedSkillRepo(t, dataDir, "https://github.com/octo/demo", "skills/demo-skill", "# Demo\nOld\n", "# Demo\nNew\n")

	sk, err := app.storage.Import(sourceDir, defaultCategoryName, skill.SourceGitHub, "https://github.com/octo/demo", "skills/demo-skill")
	require.NoError(t, err)
	sk.SourceSHA = oldSHA
	sk.LatestSHA = ""
	require.NoError(t, app.storage.UpdateMeta(sk))

	require.NoError(t, app.CheckUpdates())

	updated, err := app.storage.Get(sk.ID)
	require.NoError(t, err)
	assert.Equal(t, newSHA, updated.LatestSHA)
	assert.Equal(t, oldSHA, updated.SourceSHA)
	assert.False(t, updated.LastCheckedAt.IsZero())
}

func TestUpdateSkillRefreshesExistingPushedCopiesFromLocalCache(t *testing.T) {
	app, codexPushDir, claudePushDir, dataDir := newUpdateSkillTestApp(t)
	sourceDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nOld\n")
	_, oldSHA, newSHA := seedCachedSkillRepo(t, dataDir, "https://github.com/octo/demo", "skills/demo-skill", "# Demo\nOld\n", "# Demo\nNew\n")

	sk, err := app.storage.Import(sourceDir, defaultCategoryName, skill.SourceGitHub, "https://github.com/octo/demo", "skills/demo-skill")
	require.NoError(t, err)
	sk.SourceSHA = oldSHA
	sk.LatestSHA = newSHA
	require.NoError(t, app.storage.UpdateMeta(sk))

	app.autoPushImportedSkills("test.setup", []*skill.Skill{sk})
	conflicts, err := app.PushToTools([]string{sk.ID}, []string{"claude-code"})
	require.NoError(t, err)
	require.Empty(t, conflicts)

	codexSkillPath := filepath.Join(codexPushDir, "demo-skill", "skill.md")
	claudeSkillPath := filepath.Join(claudePushDir, "demo-skill", "skill.md")
	assertFileContentEquals(t, filepath.Join(sk.Path, "skill.md"), "# Demo\nOld\n")
	assertFileContentEquals(t, codexSkillPath, "# Demo\nOld\n")
	assertFileContentEquals(t, claudeSkillPath, "# Demo\nOld\n")

	require.NoError(t, app.UpdateSkill(sk.ID))

	assertFileContentEquals(t, filepath.Join(sk.Path, "skill.md"), "# Demo\nNew\n")
	assertFileContentEquals(t, codexSkillPath, "# Demo\nNew\n")
	assertFileContentEquals(t, claudeSkillPath, "# Demo\nNew\n")

	updated, err := app.storage.Get(sk.ID)
	require.NoError(t, err)
	assert.Equal(t, newSHA, updated.SourceSHA)
	assert.Empty(t, updated.LatestSHA)
}

func TestUpdateSkillFailsWhenLocalCacheMissing(t *testing.T) {
	app, _, _, _ := newUpdateSkillTestApp(t)
	sourceDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nOld\n")

	sk, err := app.storage.Import(sourceDir, defaultCategoryName, skill.SourceGitHub, "https://github.com/octo/demo", "skills/demo-skill")
	require.NoError(t, err)

	err = app.UpdateSkill(sk.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "local cache missing")
}

func newUpdateSkillTestApp(t *testing.T) (*App, string, string, string) {
	t.Helper()

	dataDir := t.TempDir()
	codexPushDir := filepath.Join(dataDir, "codex-skills")
	claudePushDir := filepath.Join(dataDir, "claude-skills")
	skillsDir := filepath.Join(dataDir, "library", "skills")

	svc := config.NewService(dataDir)
	cfg := config.DefaultConfig(dataDir)
	cfg.SkillsStorageDir = skillsDir
	cfg.AutoPushTools = []string{"codex"}
	cfg.Tools = []config.ToolConfig{
		{
			Name:     "codex",
			ScanDirs: []string{codexPushDir},
			PushDir:  codexPushDir,
			Enabled:  true,
		},
		{
			Name:     "claude-code",
			ScanDirs: []string{claudePushDir},
			PushDir:  claudePushDir,
			Enabled:  true,
		},
	}
	require.NoError(t, svc.Save(cfg))

	app := NewApp()
	app.config = svc
	app.storage = skill.NewStorage(skillsDir)
	app.cacheDir = filepath.Join(dataDir, "cache")
	return app, codexPushDir, claudePushDir, dataDir
}

func seedCachedSkillRepo(t *testing.T, dataDir, repoURL, skillSubPath, oldContent, newContent string) (string, string, string) {
	t.Helper()
	requireGitAvailable(t)

	repoDir, err := coregit.CacheDir(dataDir, repoURL)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(repoDir, 0755))

	runGitCmd(t, repoDir, "init")
	runGitCmd(t, repoDir, "config", "user.name", "SkillFlow Tests")
	runGitCmd(t, repoDir, "config", "user.email", "tests@skillflow.local")

	writeCachedSkillFiles(t, repoDir, skillSubPath, oldContent)
	runGitCmd(t, repoDir, "add", ".")
	runGitCmd(t, repoDir, "commit", "-m", "initial cache")
	oldSHA, err := coregit.GetSubPathSHA(context.Background(), repoDir, skillSubPath)
	require.NoError(t, err)

	writeCachedSkillFiles(t, repoDir, skillSubPath, newContent)
	runGitCmd(t, repoDir, "add", ".")
	runGitCmd(t, repoDir, "commit", "-m", "update cache")
	newSHA, err := coregit.GetSubPathSHA(context.Background(), repoDir, skillSubPath)
	require.NoError(t, err)

	return repoDir, oldSHA, newSHA
}

func writeCachedSkillFiles(t *testing.T, repoDir, skillSubPath, content string) {
	t.Helper()

	skillDir := filepath.Join(repoDir, filepath.FromSlash(skillSubPath))
	require.NoError(t, os.MkdirAll(skillDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "skill.md"), []byte(content), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "notes.txt"), []byte("cached"), 0644))
}

func requireGitAvailable(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is required for cache update tests")
	}
}

func runGitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "git %v failed: %s", args, string(output))
}

func assertFileContentEquals(t *testing.T, path string, want string) {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, want, string(data))
}
