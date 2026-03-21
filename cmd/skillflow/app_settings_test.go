package main

import (
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/config"
	platformgit "github.com/shinerio/skillflow/core/platform/git"
	sourcedomain "github.com/shinerio/skillflow/core/skillsource/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveConfigRebuildsRepoCacheConsumers(t *testing.T) {
	dataDir := t.TempDir()
	svc := config.NewService(dataDir)
	cfg := config.DefaultConfig(dataDir)
	require.NoError(t, svc.Save(cfg))

	app := NewApp()
	app.config = svc
	app.cacheDir = filepath.Join(dataDir, "cache")
	app.rebuildPathBoundServices(cfg.RepoCacheDir)

	repoURL := "https://github.com/octo/demo.git"
	oldLocalDir, err := platformgit.CacheDir(cfg.RepoCacheDir, repoURL)
	require.NoError(t, err)
	require.NoError(t, app.starStorage.Save([]sourcedomain.StarRepo{{
		URL:      repoURL,
		Name:     "octo/demo",
		Source:   "github.com/octo/demo",
		LocalDir: oldLocalDir,
	}}))

	nextCfg := cfg
	nextCfg.RepoCacheDir = filepath.Join(dataDir, "external-cache", "repos")
	require.NoError(t, app.SaveConfig(nextCfg))

	repos, err := app.starStorage.Load()
	require.NoError(t, err)
	require.Len(t, repos, 1)

	expectedLocalDir, err := platformgit.CacheDir(nextCfg.RepoCacheDir, repoURL)
	require.NoError(t, err)
	assert.Equal(t, expectedLocalDir, repos[0].LocalDir)
}
