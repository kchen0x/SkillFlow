package repository_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/promptcatalog/domain"
	"github.com/shinerio/skillflow/core/promptcatalog/infra/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilesystemStorageMigratesLegacyLayout(t *testing.T) {
	root := filepath.Join(t.TempDir(), "prompts")
	legacyDir := filepath.Join(root, "20260308-review-api")
	require.NoError(t, os.MkdirAll(legacyDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(legacyDir, domain.FileName), []byte("Legacy content"), 0644))

	store := repository.NewFilesystemStorage(root)
	items, err := store.ListAll()
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, domain.DefaultCategoryName, items[0].Category)
	assert.Equal(t, "20260308-review-api", items[0].Name)
	_, err = os.Stat(filepath.Join(root, domain.DefaultCategoryName, "20260308-review-api", domain.FileName))
	assert.NoError(t, err)
}

func TestFilesystemStorageUpdateAllowsCaseOnlyRename(t *testing.T) {
	root := filepath.Join(t.TempDir(), "prompts")
	store := repository.NewFilesystemStorage(root)

	_, err := store.Create("Gitacp", "Git helper", domain.DefaultCategoryName, "git add && git commit && git push", nil, nil)
	require.NoError(t, err)

	updated, err := store.Update("Gitacp", "gitacp", "Git helper", domain.DefaultCategoryName, "git add && git commit && git push", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "gitacp", updated.Name)
	assert.Equal(t, filepath.Join(root, domain.DefaultCategoryName, "gitacp"), updated.Path)

	items, err := store.ListAll()
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "gitacp", items[0].Name)
}
