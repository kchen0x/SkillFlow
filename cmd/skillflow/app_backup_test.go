package main

import (
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackupProfileUsesConfigServiceDataDirOnly(t *testing.T) {
	dataDir := t.TempDir()

	svc := config.NewService(dataDir)
	cfg := config.DefaultConfig(dataDir)
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
}
