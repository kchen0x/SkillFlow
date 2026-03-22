package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/config"
	"github.com/shinerio/skillflow/core/platform/appdata"
	platformgit "github.com/shinerio/skillflow/core/platform/git"
	skillcatalogapp "github.com/shinerio/skillflow/core/skillcatalog/app"
	skilldomain "github.com/shinerio/skillflow/core/skillcatalog/domain"
	skillrepo "github.com/shinerio/skillflow/core/skillcatalog/infra/repository"
	sourcedomain "github.com/shinerio/skillflow/core/skillsource/domain"
	sourcerepo "github.com/shinerio/skillflow/core/skillsource/infra/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleRestoredCloudStateAutoPushesNewlyRestoredSkills(t *testing.T) {
	app, pushDir, _, _ := newRestoreTestApp(t, []string{"codex"})
	before := app.captureCloudRestoreState()
	sourceDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nRestored\n")

	_, err := app.storage.Import(sourceDir, defaultCategoryName, skilldomain.SourceManual, "", "")
	require.NoError(t, err)

	require.NoError(t, app.handleRestoredCloudState(before, "test.restore"))

	pushedPath := filepath.Join(pushDir, "demo-skill", "skill.md")
	assert.FileExists(t, pushedPath)

	pushedContent, err := os.ReadFile(pushedPath)
	require.NoError(t, err)
	assert.Equal(t, "# Demo\nRestored\n", string(pushedContent))
}

func TestHandleRestoredCloudStateAutoPushesUpdatedExistingSkills(t *testing.T) {
	app, pushDir, _, _ := newRestoreTestApp(t, []string{"codex"})
	sourceDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nVersion 1\n")

	sk, err := app.storage.Import(sourceDir, defaultCategoryName, skilldomain.SourceManual, "", "")
	require.NoError(t, err)

	app.autoPushImportedSkillsToAgents("test.setup", []*skilldomain.InstalledSkill{sk})
	pushPath := filepath.Join(pushDir, "demo-skill", "skill.md")
	assertFileContentEquals(t, pushPath, "# Demo\nVersion 1\n")

	before := app.captureCloudRestoreState()
	require.NoError(t, os.WriteFile(filepath.Join(sk.Path, "skill.md"), []byte("# Demo\nVersion 2\n"), 0644))
	require.NoError(t, app.storage.UpdateMeta(sk))

	require.NoError(t, app.handleRestoredCloudState(before, "test.restore"))
	assertFileContentEquals(t, pushPath, "# Demo\nVersion 2\n")
}

func TestHandleRestoredCloudStateClonesNewlyRestoredStarredRepos(t *testing.T) {
	if err := platformgit.CheckGitInstalled(); err != nil {
		t.Skip("git not installed")
	}

	app, _, _, _ := newRestoreTestApp(t, nil)
	before := app.captureCloudRestoreState()
	sourceRepo := newLocalGitRepo(t)
	repoURL := "https://example.com/restored/test.git"
	cloneDir, err := platformgit.CacheDir(app.repoCacheDir(), repoURL)
	require.NoError(t, err)

	prevCloneOrUpdateRepo := cloneOrUpdateRepo
	cloneOrUpdateRepo = func(ctx context.Context, gotRepoURL, dir, proxyURL string) error {
		if gotRepoURL != repoURL {
			return exec.ErrNotFound
		}
		return platformgit.CloneOrUpdate(ctx, sourceRepo, dir, proxyURL)
	}
	t.Cleanup(func() {
		cloneOrUpdateRepo = prevCloneOrUpdateRepo
	})

	require.NoError(t, app.starStorage.Save([]sourcedomain.StarRepo{{
		URL:      repoURL,
		Name:     "local/test-repo",
		Source:   "example.com/restored/test",
		LocalDir: cloneDir,
	}}))

	require.NoError(t, app.handleRestoredCloudState(before, "test.restore"))

	assert.DirExists(t, filepath.Join(cloneDir, ".git"))
	assert.FileExists(t, filepath.Join(cloneDir, "README.md"))

	repos, err := app.starStorage.Load()
	require.NoError(t, err)
	require.Len(t, repos, 1)
	assert.Empty(t, repos[0].SyncError)
	assert.False(t, repos[0].LastSync.IsZero())
}

func newRestoreTestApp(t *testing.T, autoPushAgents []string) (*App, string, string, string) {
	t.Helper()

	dataDir := t.TempDir()
	pushDir := filepath.Join(dataDir, "agent-skills")
	skillsDir := appdata.SkillsDir(dataDir)
	starsPath := filepath.Join(dataDir, "star_repos.json")
	require.NoError(t, os.WriteFile(starsPath, []byte("[]"), 0644))

	svc := config.NewService(dataDir)
	cfg := config.DefaultConfig(dataDir)
	cfg.AutoPushAgents = autoPushAgents
	cfg.Agents = []config.AgentConfig{{
		Name:     "codex",
		ScanDirs: []string{pushDir},
		PushDir:  pushDir,
		Enabled:  true,
	}}
	require.NoError(t, svc.Save(cfg))

	app := NewApp()
	app.config = svc
	app.storage = skillcatalogapp.NewService(skillrepo.NewFilesystemStorage(skillsDir))
	app.cacheDir = filepath.Join(dataDir, "cache")
	app.starStorage = sourcerepo.NewStarRepoStorageWithCacheDir(starsPath, cfg.RepoCacheDir)
	return app, pushDir, skillsDir, starsPath
}

func newLocalGitRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@test.com")
	runGit(t, dir, "config", "user.name", "Test")
	runGit(t, dir, "config", "commit.gpgsign", "false")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello"), 0644))
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "init")
	return dir
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", append([]string{"-c", "commit.gpgsign=false"}, args...)...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v failed: %s", args, string(output))
}
