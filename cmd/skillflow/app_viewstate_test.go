package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/config"
	coregit "github.com/shinerio/skillflow/core/git"
	"github.com/shinerio/skillflow/core/skill"
	"github.com/shinerio/skillflow/core/viewstate"
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
	cached := []coregit.StarSkill{{Name: "cached-skill", Path: "/tmp/cached"}}
	require.NoError(t, app.viewCache.Save(allStarSkillsSnapshotName, fingerprint, cached))

	skills, err := app.ListAllStarSkills()
	require.NoError(t, err)
	assert.Equal(t, cached, skills)
}

func TestListAllStarSkillsRebuildsWhenCachedSnapshotIsStale(t *testing.T) {
	app, dataDir, _, _ := newViewStateTestApp(t)
	repoDir := filepath.Join(dataDir, "repos", "demo")
	writeTestSkillDir(t, repoDir, "repo-skill", "# Repo Skill\n")
	require.NoError(t, app.starStorage.Save([]coregit.StarredRepo{{
		URL:      "https://example.com/demo.git",
		Name:     "demo/repo",
		Source:   "demo/repo",
		LocalDir: repoDir,
	}}))

	require.NoError(t, app.viewCache.Save(allStarSkillsSnapshotName, "stale-fingerprint", []coregit.StarSkill{{
		Name: "cached-skill",
		Path: "/tmp/cached",
	}}))

	skills, err := app.ListAllStarSkills()
	require.NoError(t, err)
	require.Len(t, skills, 1)
	assert.Equal(t, "repo-skill", skills[0].Name)
	assert.Equal(t, filepath.Join(repoDir, "repo-skill"), skills[0].Path)
}

func newViewStateTestApp(t *testing.T) (*App, string, string, string) {
	t.Helper()

	dataDir := t.TempDir()
	skillsDir := filepath.Join(dataDir, "library", "skills")
	pushDir := filepath.Join(dataDir, "tool-skills")
	cacheDir := filepath.Join(dataDir, "cache")
	starsPath := filepath.Join(dataDir, "star_repos.json")

	svc := config.NewService(dataDir)
	cfg := config.DefaultConfig(dataDir)
	cfg.SkillsStorageDir = skillsDir
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
	app.storage = skill.NewStorage(skillsDir)
	app.cacheDir = cacheDir
	app.viewCache = viewstate.NewManager(filepath.Join(cacheDir, "viewstate"))
	app.starStorage = coregit.NewStarStorage(starsPath)
	return app, dataDir, skillsDir, pushDir
}
