package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/config"
	platformgit "github.com/shinerio/skillflow/core/platform/git"
	skillcatalogapp "github.com/shinerio/skillflow/core/skillcatalog/app"
	skilldomain "github.com/shinerio/skillflow/core/skillcatalog/domain"
	skillrepo "github.com/shinerio/skillflow/core/skillcatalog/infra/repository"
	sourcedomain "github.com/shinerio/skillflow/core/skillsource/domain"
	sourcerepo "github.com/shinerio/skillflow/core/skillsource/infra/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckUpdatesUsesLocalCacheSHA(t *testing.T) {
	app, _, _, dataDir := newUpdateSkillTestApp(t)
	sourceDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nOld\n")
	_, oldSHA, newSHA := seedCachedSkillRepo(t, dataDir, "https://github.com/octo/demo", "skills/demo-skill", "# Demo\nOld\n", "# Demo\nNew\n")

	sk, err := app.storage.Import(sourceDir, defaultCategoryName, skilldomain.SourceGitHub, "https://github.com/octo/demo", "skills/demo-skill")
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

	sk, err := app.storage.Import(sourceDir, defaultCategoryName, skilldomain.SourceGitHub, "https://github.com/octo/demo", "skills/demo-skill")
	require.NoError(t, err)
	sk.SourceSHA = oldSHA
	sk.LatestSHA = newSHA
	require.NoError(t, app.storage.UpdateMeta(sk))

	app.autoPushImportedSkillsToAgents("test.setup", []*skilldomain.InstalledSkill{sk})
	conflicts, err := app.PushToAgents([]string{sk.ID}, []string{"claude-code"})
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

func TestUpdateSkillAutoPushesToSelectedAgentsWhenMissing(t *testing.T) {
	app, codexPushDir, claudePushDir, dataDir := newUpdateSkillTestApp(t)
	sourceDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nOld\n")
	_, oldSHA, newSHA := seedCachedSkillRepo(t, dataDir, "https://github.com/octo/demo", "skills/demo-skill", "# Demo\nOld\n", "# Demo\nNew\n")

	sk, err := app.storage.Import(sourceDir, defaultCategoryName, skilldomain.SourceGitHub, "https://github.com/octo/demo", "skills/demo-skill")
	require.NoError(t, err)
	sk.SourceSHA = oldSHA
	sk.LatestSHA = newSHA
	require.NoError(t, app.storage.UpdateMeta(sk))

	require.NoError(t, app.UpdateSkill(sk.ID))

	assertFileContentEquals(t, filepath.Join(codexPushDir, "demo-skill", "skill.md"), "# Demo\nNew\n")
	_, err = os.Stat(filepath.Join(claudePushDir, "demo-skill"))
	assert.True(t, os.IsNotExist(err))
}

func TestUpdateSkillAutoPushOverwritesSelectedAgentSkill(t *testing.T) {
	app, codexPushDir, _, dataDir := newUpdateSkillTestApp(t)
	sourceDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nOld\n")
	_, oldSHA, newSHA := seedCachedSkillRepo(t, dataDir, "https://github.com/octo/demo", "skills/demo-skill", "# Demo\nOld\n", "# Demo\nNew\n")

	sk, err := app.storage.Import(sourceDir, defaultCategoryName, skilldomain.SourceGitHub, "https://github.com/octo/demo", "skills/demo-skill")
	require.NoError(t, err)
	sk.SourceSHA = oldSHA
	sk.LatestSHA = newSHA
	require.NoError(t, app.storage.UpdateMeta(sk))

	agentSkillDir := filepath.Join(codexPushDir, "demo-skill")
	require.NoError(t, os.MkdirAll(agentSkillDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(agentSkillDir, "skill.md"), []byte("# Demo\nOld Agent Copy\n"), 0644))

	require.NoError(t, app.UpdateSkill(sk.ID))
	assertFileContentEquals(t, filepath.Join(agentSkillDir, "skill.md"), "# Demo\nNew\n")
}

func TestUpdateSkillFailsWhenLocalCacheMissing(t *testing.T) {
	app, _, _, _ := newUpdateSkillTestApp(t)
	sourceDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nOld\n")

	sk, err := app.storage.Import(sourceDir, defaultCategoryName, skilldomain.SourceGitHub, "https://github.com/octo/demo", "skills/demo-skill")
	require.NoError(t, err)

	err = app.UpdateSkill(sk.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "local cache missing")
}

func TestUpdateStarredRepoAutoUpdatesMatchingInstalledSkillsWhenEnabled(t *testing.T) {
	app, codexPushDir, _, dataDir := newUpdateSkillTestApp(t)
	setAutoUpdateSkills(t, app, true)

	fixture := setupStarredRepoAutoUpdateFixture(t, app, dataDir, "https://github.com/octo/demo", "skills/demo-skill", "# Demo\nOld\n")
	newSHA := commitLocalRemoteSkillRepoUpdate(t, fixture.remoteRepoDir, fixture.skillSubPath, "# Demo\nNew\n")
	stubCloneOrUpdateRepo(t, fixture.repoURL, fixture.remoteRepoDir)

	require.NoError(t, app.UpdateStarredRepo(fixture.repoURL))

	assertFileContentEquals(t, filepath.Join(fixture.skill.Path, "skill.md"), "# Demo\nNew\n")
	assertFileContentEquals(t, filepath.Join(codexPushDir, "demo-skill", "skill.md"), "# Demo\nNew\n")

	updated, err := app.storage.Get(fixture.skill.ID)
	require.NoError(t, err)
	assert.Equal(t, newSHA, updated.SourceSHA)
	assert.Empty(t, updated.LatestSHA)
}

func TestUpdateStarredRepoDoesNotAutoUpdateWhenDisabled(t *testing.T) {
	app, codexPushDir, _, dataDir := newUpdateSkillTestApp(t)
	setAutoUpdateSkills(t, app, false)

	fixture := setupStarredRepoAutoUpdateFixture(t, app, dataDir, "https://github.com/octo/demo", "skills/demo-skill", "# Demo\nOld\n")
	commitLocalRemoteSkillRepoUpdate(t, fixture.remoteRepoDir, fixture.skillSubPath, "# Demo\nNew\n")
	stubCloneOrUpdateRepo(t, fixture.repoURL, fixture.remoteRepoDir)

	require.NoError(t, app.UpdateStarredRepo(fixture.repoURL))

	assertFileContentEquals(t, filepath.Join(fixture.skill.Path, "skill.md"), "# Demo\nOld\n")
	_, err := os.Stat(filepath.Join(codexPushDir, "demo-skill"))
	assert.True(t, os.IsNotExist(err))

	updated, err := app.storage.Get(fixture.skill.ID)
	require.NoError(t, err)
	assert.Equal(t, fixture.oldSHA, updated.SourceSHA)
}

func TestUpdateAllStarredReposAutoUpdatesOnlySuccessfulRepos(t *testing.T) {
	app, _, _, dataDir := newUpdateSkillTestApp(t)
	setAutoUpdateSkills(t, app, true)

	good := setupStarredRepoAutoUpdateFixture(t, app, dataDir, "https://github.com/octo/demo", "skills/demo-skill", "# Demo\nOld\n")
	bad := setupStarredRepoAutoUpdateFixture(t, app, dataDir, "https://github.com/octo/other", "skills/other-skill", "# Other\nOld\n")

	goodNewSHA := commitLocalRemoteSkillRepoUpdate(t, good.remoteRepoDir, good.skillSubPath, "# Demo\nNew\n")
	commitLocalRemoteSkillRepoUpdate(t, bad.remoteRepoDir, bad.skillSubPath, "# Other\nNew\n")

	prevCloneOrUpdateRepo := cloneOrUpdateRepo
	cloneOrUpdateRepo = func(ctx context.Context, repoURL, dir, proxyURL string) error {
		switch {
		case platformgit.SameRepo(repoURL, good.repoURL):
			return platformgit.CloneOrUpdate(ctx, good.remoteRepoDir, dir, proxyURL)
		case platformgit.SameRepo(repoURL, bad.repoURL):
			return fmt.Errorf("forced sync failure for %s", repoURL)
		default:
			return fmt.Errorf("unexpected repo url: %s", repoURL)
		}
	}
	t.Cleanup(func() {
		cloneOrUpdateRepo = prevCloneOrUpdateRepo
	})

	require.NoError(t, app.UpdateAllStarredRepos())

	assertFileContentEquals(t, filepath.Join(good.skill.Path, "skill.md"), "# Demo\nNew\n")
	assertFileContentEquals(t, filepath.Join(bad.skill.Path, "skill.md"), "# Other\nOld\n")

	goodUpdated, err := app.storage.Get(good.skill.ID)
	require.NoError(t, err)
	assert.Equal(t, goodNewSHA, goodUpdated.SourceSHA)

	badUpdated, err := app.storage.Get(bad.skill.ID)
	require.NoError(t, err)
	assert.Equal(t, bad.oldSHA, badUpdated.SourceSHA)
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
	cfg.AutoPushAgents = []string{"codex"}
	cfg.Agents = []config.AgentConfig{
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
	app.storage = skillcatalogapp.NewService(skillrepo.NewFilesystemStorage(skillsDir))
	app.cacheDir = filepath.Join(dataDir, "cache")
	return app, codexPushDir, claudePushDir, dataDir
}

type starredRepoAutoUpdateFixture struct {
	repoURL       string
	skillSubPath  string
	remoteRepoDir string
	skill         *skilldomain.InstalledSkill
	oldSHA        string
}

func setAutoUpdateSkills(t *testing.T, app *App, enabled bool) {
	t.Helper()

	cfg, err := app.config.Load()
	require.NoError(t, err)
	cfg.AutoUpdateSkills = enabled
	require.NoError(t, app.config.Save(cfg))
}

func setupStarredRepoAutoUpdateFixture(t *testing.T, app *App, dataDir, repoURL, skillSubPath, oldContent string) starredRepoAutoUpdateFixture {
	t.Helper()

	remoteRepoDir := t.TempDir()
	oldSHA := seedLocalRemoteSkillRepo(t, remoteRepoDir, skillSubPath, oldContent)

	cacheDir, err := platformgit.CacheDir(dataDir, repoURL)
	require.NoError(t, err)
	require.NoError(t, platformgit.CloneOrUpdate(context.Background(), remoteRepoDir, cacheDir, ""))

	sourceDir := writeTestSkillDir(t, t.TempDir(), filepath.Base(filepath.FromSlash(skillSubPath)), oldContent)
	sk, err := app.storage.Import(sourceDir, defaultCategoryName, skilldomain.SourceGitHub, repoURL, skillSubPath)
	require.NoError(t, err)
	sk.SourceSHA = oldSHA
	sk.LatestSHA = ""
	require.NoError(t, app.storage.UpdateMeta(sk))

	if app.starStorage == nil {
		app.starStorage = sourcerepo.NewStarRepoStorage(filepath.Join(dataDir, "star_repos.json"))
	}
	repos, err := app.starStorage.Load()
	require.NoError(t, err)
	repos = append(repos, sourcedomain.StarRepo{
		URL:      repoURL,
		Name:     "octo/" + filepath.Base(filepath.FromSlash(skillSubPath)),
		LocalDir: cacheDir,
	})
	require.NoError(t, app.starStorage.Save(repos))

	return starredRepoAutoUpdateFixture{
		repoURL:       repoURL,
		skillSubPath:  skillSubPath,
		remoteRepoDir: remoteRepoDir,
		skill:         sk,
		oldSHA:        oldSHA,
	}
}

func seedCachedSkillRepo(t *testing.T, dataDir, repoURL, skillSubPath, oldContent, newContent string) (string, string, string) {
	t.Helper()
	requireGitAvailable(t)

	repoDir, err := platformgit.CacheDir(dataDir, repoURL)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(repoDir, 0755))

	runGitCmd(t, repoDir, "init")
	runGitCmd(t, repoDir, "config", "user.name", "SkillFlow Tests")
	runGitCmd(t, repoDir, "config", "user.email", "tests@skillflow.local")
	runGitCmd(t, repoDir, "config", "commit.gpgsign", "false")

	writeCachedSkillFiles(t, repoDir, skillSubPath, oldContent)
	runGitCmd(t, repoDir, "add", ".")
	runGitCmd(t, repoDir, "commit", "-m", "initial cache")
	oldSHA, err := platformgit.GetSubPathSHA(context.Background(), repoDir, skillSubPath)
	require.NoError(t, err)

	writeCachedSkillFiles(t, repoDir, skillSubPath, newContent)
	runGitCmd(t, repoDir, "add", ".")
	runGitCmd(t, repoDir, "commit", "-m", "update cache")
	newSHA, err := platformgit.GetSubPathSHA(context.Background(), repoDir, skillSubPath)
	require.NoError(t, err)

	return repoDir, oldSHA, newSHA
}

func seedLocalRemoteSkillRepo(t *testing.T, repoDir, skillSubPath, content string) string {
	t.Helper()
	requireGitAvailable(t)

	require.NoError(t, os.MkdirAll(repoDir, 0755))
	runGitCmd(t, repoDir, "init")
	runGitCmd(t, repoDir, "config", "user.name", "SkillFlow Tests")
	runGitCmd(t, repoDir, "config", "user.email", "tests@skillflow.local")
	runGitCmd(t, repoDir, "config", "commit.gpgsign", "false")
	writeCachedSkillFiles(t, repoDir, skillSubPath, content)
	runGitCmd(t, repoDir, "add", ".")
	runGitCmd(t, repoDir, "commit", "-m", "initial remote")

	sha, err := platformgit.GetSubPathSHA(context.Background(), repoDir, skillSubPath)
	require.NoError(t, err)
	return sha
}

func commitLocalRemoteSkillRepoUpdate(t *testing.T, repoDir, skillSubPath, content string) string {
	t.Helper()

	writeCachedSkillFiles(t, repoDir, skillSubPath, content)
	runGitCmd(t, repoDir, "add", ".")
	runGitCmd(t, repoDir, "commit", "-m", "update remote")

	sha, err := platformgit.GetSubPathSHA(context.Background(), repoDir, skillSubPath)
	require.NoError(t, err)
	return sha
}

func stubCloneOrUpdateRepo(t *testing.T, wantRepoURL, sourceRepoDir string) {
	t.Helper()

	prevCloneOrUpdateRepo := cloneOrUpdateRepo
	cloneOrUpdateRepo = func(ctx context.Context, repoURL, dir, proxyURL string) error {
		if !platformgit.SameRepo(repoURL, wantRepoURL) {
			return fmt.Errorf("unexpected repo url: %s", repoURL)
		}
		return platformgit.CloneOrUpdate(ctx, sourceRepoDir, dir, proxyURL)
	}
	t.Cleanup(func() {
		cloneOrUpdateRepo = prevCloneOrUpdateRepo
	})
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
