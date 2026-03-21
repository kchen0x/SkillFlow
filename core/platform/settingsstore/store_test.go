package settingsstore_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/platform/settingsstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStoreUsesStandardConfigPaths(t *testing.T) {
	root := t.TempDir()

	store := settingsstore.New(root)

	assert.Equal(t, root, store.DataDir())
	assert.Equal(t, filepath.Join(root, "config.json"), store.SharedPath())
	assert.Equal(t, filepath.Join(root, "config_local.json"), store.LocalPath())
}

func TestStoreWritesAndReadsSharedAndLocalJSON(t *testing.T) {
	root := t.TempDir()
	store := settingsstore.New(root)

	type sharedDoc struct {
		Name string `json:"name"`
	}
	type localDoc struct {
		Path string `json:"path"`
	}

	require.NoError(t, store.WriteShared(sharedDoc{Name: "shared"}))
	require.NoError(t, store.WriteLocal(localDoc{Path: "/tmp/skillflow"}))

	var shared sharedDoc
	exists, err := store.ReadShared(&shared)
	require.NoError(t, err)
	require.True(t, exists)
	assert.Equal(t, "shared", shared.Name)

	var local localDoc
	exists, err = store.ReadLocal(&local)
	require.NoError(t, err)
	require.True(t, exists)
	assert.Equal(t, "/tmp/skillflow", local.Path)
}

func TestStoreReturnsMissingWithoutError(t *testing.T) {
	store := settingsstore.New(t.TempDir())

	var shared map[string]any
	exists, err := store.ReadShared(&shared)
	require.NoError(t, err)
	assert.False(t, exists)

	var local map[string]any
	exists, err = store.ReadLocal(&local)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestNormalizeWindowStateRejectsSmallDimensions(t *testing.T) {
	assert.Equal(t, settingsstore.WindowState{}, settingsstore.NormalizeWindowState(settingsstore.WindowState{Width: 320, Height: 920}))
	assert.Equal(t, settingsstore.WindowState{}, settingsstore.NormalizeWindowState(settingsstore.WindowState{Width: 1440, Height: 300}))
	assert.Equal(t, settingsstore.WindowState{Width: 1440, Height: 920}, settingsstore.NormalizeWindowState(settingsstore.WindowState{Width: 1440, Height: 920}))
}

func TestStoreSavesAndLoadsWindowStateInLocalConfig(t *testing.T) {
	root := t.TempDir()
	store := settingsstore.New(root)

	require.NoError(t, store.SaveWindowState(settingsstore.WindowState{Width: 1440, Height: 920}))

	state, ok, err := store.LoadWindowState()
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, settingsstore.WindowState{Width: 1440, Height: 920}, state)

	sharedData, err := os.ReadFile(store.SharedPath())
	if err == nil {
		assert.NotContains(t, string(sharedData), `"window"`)
	} else {
		assert.ErrorIs(t, err, os.ErrNotExist)
	}

	localData, err := os.ReadFile(store.LocalPath())
	require.NoError(t, err)
	assert.Contains(t, string(localData), `"window"`)
	assert.Contains(t, string(localData), `"width": 1440`)
	assert.Contains(t, string(localData), `"height": 920`)
}
