package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/config"
	platformgit "github.com/shinerio/skillflow/core/platform/git"
	"github.com/shinerio/skillflow/core/platform/appdata"
	"github.com/shinerio/skillflow/core/readmodel/viewstate"
	skillcatalogapp "github.com/shinerio/skillflow/core/skillcatalog/app"
	skillrepo "github.com/shinerio/skillflow/core/skillcatalog/infra/repository"
	sourcedomain "github.com/shinerio/skillflow/core/skillsource/domain"
	sourcerepo "github.com/shinerio/skillflow/core/skillsource/infra/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMeasureOperationReturnsOperationResult(t *testing.T) {
	app := NewApp()

	value, err := measureOperation(app, "test.operation", func() (int, error) {
		return 42, nil
	})
	require.NoError(t, err)
	assert.Equal(t, 42, value)
}

func TestListSkillsUsesDaemonServiceForUIProcessRole(t *testing.T) {
	prevActiveProcessRole := activeProcessRole
	prevDaemonInvokeServiceFn := daemonInvokeServiceFn
	t.Cleanup(func() {
		activeProcessRole = prevActiveProcessRole
		daemonInvokeServiceFn = prevDaemonInvokeServiceFn
	})

	activeProcessRole = processRoleUI
	expected := []InstalledSkillEntry{
		{ID: "skill-1", Name: "Demo", Category: defaultCategoryName, Source: "manual"},
	}
	daemonInvokeServiceFn = func(method string, params any, result any) error {
		assert.Equal(t, "ListSkills", method)
		assert.Nil(t, params)
		target, ok := result.(*[]InstalledSkillEntry)
		require.True(t, ok)
		*target = expected
		return nil
	}

	app := NewApp()
	entries, err := app.ListSkills()
	require.NoError(t, err)
	assert.Equal(t, expected, entries)
}

func TestListCategoriesUsesDaemonServiceForUIProcessRole(t *testing.T) {
	prevActiveProcessRole := activeProcessRole
	prevDaemonInvokeServiceFn := daemonInvokeServiceFn
	t.Cleanup(func() {
		activeProcessRole = prevActiveProcessRole
		daemonInvokeServiceFn = prevDaemonInvokeServiceFn
	})

	activeProcessRole = processRoleUI
	expected := []string{defaultCategoryName, "Tools"}
	daemonInvokeServiceFn = func(method string, params any, result any) error {
		assert.Equal(t, "ListCategories", method)
		assert.Nil(t, params)
		target, ok := result.(*[]string)
		require.True(t, ok)
		*target = expected
		return nil
	}

	app := NewApp()
	categories, err := app.ListCategories()
	require.NoError(t, err)
	assert.Equal(t, expected, categories)
}

func TestListSkillsUsesCachedSnapshotWhenFingerprintMatches(t *testing.T) {
	app, _, _, _ := newViewStateTestApp(t)

	fingerprint, err := app.installedSkillsFingerprint()
	require.NoError(t, err)
	cached := []InstalledSkillEntry{
		{ID: "cached", Name: "Cached Skill", Category: defaultCategoryName, Source: "manual"},
	}
	require.NoError(t, app.viewCache.Save(installedSkillsSnapshotName, fingerprint, cached))

	entries, err := app.ListSkills()
	require.NoError(t, err)
	assert.Equal(t, cached, entries)
}

func TestListSkillsRebuildsWhenCachedSnapshotIsStale(t *testing.T) {
	app, _, skillsDir, _ := newViewStateTestApp(t)
	sourceDir := writeTestSkillDir(t, t.TempDir(), "live-skill", "# Live\n")

	_, err := app.ImportLocal(sourceDir, "")
	require.NoError(t, err)

	require.NoError(t, app.viewCache.Save(installedSkillsSnapshotName, "stale-fingerprint", []InstalledSkillEntry{
		{ID: "cached", Name: "Cached Skill", Category: defaultCategoryName, Source: "manual"},
	}))

	entries, err := app.ListSkills()
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "live-skill", entries[0].Name)
	assert.Equal(t, filepath.Join(skillsDir, defaultCategoryName, "live-skill"), entries[0].Path)
}

func TestListAllStarSkillsUsesCachedSnapshotWhenFingerprintMatches(t *testing.T) {
	app, _, _, _ := newViewStateTestApp(t)

	fingerprint, err := app.allStarSkillsFingerprint()
	require.NoError(t, err)
	cached := []StarSkillEntry{{Name: "cached-skill", Path: "/tmp/cached"}}
	require.NoError(t, app.viewCache.Save(allStarSkillsSnapshotName, fingerprint, cached))

	skills, err := app.ListAllStarSkills()
	require.NoError(t, err)
	assert.Equal(t, cached, skills)
}

func TestListAllStarSkillsRebuildsWhenCachedSnapshotIsStale(t *testing.T) {
	app, _, _, _ := newViewStateTestApp(t)
	repoURL := "https://example.com/demo/repo.git"
	repoDir, err := platformgit.CacheDir(app.repoCacheDir(), repoURL)
	require.NoError(t, err)
	writeTestSkillDir(t, repoDir, "repo-skill", "# Repo Skill\n")
	require.NoError(t, app.starStorage.Save([]sourcedomain.StarRepo{{
		URL:      repoURL,
		Name:     "demo/repo",
		Source:   "demo/repo",
		LocalDir: repoDir,
	}}))

	require.NoError(t, app.viewCache.Save(allStarSkillsSnapshotName, "stale-fingerprint", []StarSkillEntry{{
		Name: "cached-skill",
		Path: "/tmp/cached",
	}}))

	skills, err := app.ListAllStarSkills()
	require.NoError(t, err)
	require.Len(t, skills, 1)
	assert.Equal(t, "repo-skill", skills[0].Name)
	assert.Equal(t, filepath.Join(repoDir, "repo-skill"), skills[0].Path)
}

func TestInstalledSkillsFingerprintUsesAppDataMetaDirs(t *testing.T) {
	app, dataDir, _, _ := newViewStateTestApp(t)
	metaDir := filepath.Join(dataDir, "meta")
	metaLocalDir := filepath.Join(dataDir, "meta_local")
	require.NoError(t, os.MkdirAll(metaDir, 0755))
	require.NoError(t, os.MkdirAll(metaLocalDir, 0755))

	first, err := app.installedSkillsFingerprint()
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(metaDir, "demo.json"), []byte(`{"name":"demo"}`), 0644))
	second, err := app.installedSkillsFingerprint()
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(metaLocalDir, "demo.json"), []byte(`{"state":"local"}`), 0644))
	third, err := app.installedSkillsFingerprint()
	require.NoError(t, err)

	assert.NotEqual(t, first, second)
	assert.NotEqual(t, second, third)
}

func newViewStateTestApp(t *testing.T) (*App, string, string, string) {
	t.Helper()

	dataDir := t.TempDir()
	skillsDir := appdata.SkillsDir(dataDir)
	pushDir := filepath.Join(dataDir, "tool-skills")
	cacheDir := filepath.Join(dataDir, "cache")
	starsPath := filepath.Join(dataDir, "star_repos.json")

	svc := config.NewService(dataDir)
	cfg := config.DefaultConfig(dataDir)
	cfg.Agents = []config.AgentConfig{
		{
			Name:     "codex",
			ScanDirs: []string{pushDir},
			PushDir:  pushDir,
			Enabled:  true,
		},
	}
	require.NoError(t, svc.Save(cfg))
	require.NoError(t, os.MkdirAll(pushDir, 0755))

	app := NewApp()
	app.config = svc
	app.storage = skillcatalogapp.NewService(skillrepo.NewFilesystemStorage(skillsDir))
	app.cacheDir = cacheDir
	app.viewCache = viewstate.NewManager(filepath.Join(cacheDir, "viewstate"))
	app.starStorage = sourcerepo.NewStarRepoStorageWithCacheDir(starsPath, cfg.RepoCacheDir)
	return app, dataDir, skillsDir, pushDir
}
