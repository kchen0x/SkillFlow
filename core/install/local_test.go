package install_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/install"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalInstallerScanValidSkill(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "skill.md"), []byte("# skill"), 0644))

	inst := install.NewLocalInstaller()
	candidates, err := inst.Scan(context.Background(), install.InstallSource{Type: "local", URI: dir})
	require.NoError(t, err)
	assert.Len(t, candidates, 1)
	assert.Equal(t, filepath.Base(dir), candidates[0].Name)
}

func TestLocalInstallerScanInvalidSkill(t *testing.T) {
	dir := t.TempDir() // no skill.md
	inst := install.NewLocalInstaller()
	candidates, err := inst.Scan(context.Background(), install.InstallSource{Type: "local", URI: dir})
	require.NoError(t, err)
	assert.Empty(t, candidates)
}
