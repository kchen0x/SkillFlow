package main

import (
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackupProfileUsesConfigServiceDataDir(t *testing.T) {
	dataDir := t.TempDir()
	skillsDir := filepath.Join(dataDir, "library", "skills")

	svc := config.NewService(dataDir)
	cfg := config.DefaultConfig(dataDir)
	cfg.SkillsStorageDir = skillsDir
	require.NoError(t, svc.Save(cfg))

	prevAppDataDirFunc := appDataDirFunc
	t.Cleanup(func() {
		appDataDirFunc = prevAppDataDirFunc
	})
	appDataDirFunc = func() string {
		return filepath.Join(t.TempDir(), "wrong-app-data-dir")
	}

	app := NewApp()
	app.config = svc

	profile := app.backupProfile(cfg)
	assert.Equal(t, dataDir, profile.AppDataDir)
	assert.Equal(t, skillsDir, profile.SkillsStorageDir)
}
