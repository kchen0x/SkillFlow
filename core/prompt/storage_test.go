package prompt_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/prompt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorageCreateUpdateMoveDelete(t *testing.T) {
	root := filepath.Join(t.TempDir(), "prompts")
	store := prompt.NewStorage(root)

	created, err := store.Create("Review API", "Review backend changes", "Default", "Please review the API diff.")
	require.NoError(t, err)
	assert.Equal(t, "Review API", created.Name)
	assert.Equal(t, "Default", created.Category)
	assert.Equal(t, filepath.Join(root, "Default", "Review API", prompt.FileName), created.FilePath)

	updated, err := store.Update("Review API", "Review API", "Review backend and frontend changes", "Writing", "Please review the UI diff too.")
	require.NoError(t, err)
	assert.Equal(t, "Writing", updated.Category)
	assert.Equal(t, "Review backend and frontend changes", updated.Description)
	assert.Equal(t, "Please review the UI diff too.", updated.Content)

	items, err := store.ListAll()
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "Writing", items[0].Category)

	require.NoError(t, store.Delete("Review API"))
	_, err = os.Stat(filepath.Join(root, "Writing", "Review API"))
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestStorageCategories(t *testing.T) {
	root := filepath.Join(t.TempDir(), "prompts")
	store := prompt.NewStorage(root)
	require.NoError(t, store.CreateCategory("Default"))
	require.NoError(t, store.CreateCategory("Research"))

	categories, err := store.ListCategories()
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"Default", "Research"}, categories)

	_, err = store.Create("Architecture", "", "Research", "Summarize the architecture.")
	require.NoError(t, err)

	err = store.DeleteCategory("Research")
	assert.ErrorIs(t, err, prompt.ErrCategoryNotEmpty)

	require.NoError(t, store.MoveCategory("Architecture", "Default"))
	require.NoError(t, store.DeleteCategory("Research"))
}

func TestStorageExportImportJSON(t *testing.T) {
	root := filepath.Join(t.TempDir(), "prompts")
	store := prompt.NewStorage(root)
	_, err := store.Create("Prompt A", "Desc A", "Default", "Content A")
	require.NoError(t, err)
	_, err = store.Create("Prompt B", "Desc B", "Ops", "Content B")
	require.NoError(t, err)

	data, err := store.ExportJSON()
	require.NoError(t, err)
	assert.Contains(t, string(data), "Prompt A")
	assert.Contains(t, string(data), "Prompt B")

	other := prompt.NewStorage(filepath.Join(t.TempDir(), "prompts"))
	count, err := other.ImportJSON(data)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	items, err := other.ListAll()
	require.NoError(t, err)
	require.Len(t, items, 2)
}

func TestStorageImportArrayPayloadUpdatesExistingPrompt(t *testing.T) {
	root := filepath.Join(t.TempDir(), "prompts")
	store := prompt.NewStorage(root)
	_, err := store.Create("Prompt A", "Old", "Default", "Old content")
	require.NoError(t, err)

	payload, err := json.Marshal([]map[string]string{{
		"name":        "Prompt A",
		"description": "New",
		"category":    "Research",
		"content":     "New content",
	}})
	require.NoError(t, err)

	count, err := store.ImportJSON(payload)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	item, err := store.Get("Prompt A")
	require.NoError(t, err)
	assert.Equal(t, "Research", item.Category)
	assert.Equal(t, "New", item.Description)
	assert.Equal(t, "New content", item.Content)
}

func TestStorageMigratesLegacyLayout(t *testing.T) {
	root := filepath.Join(t.TempDir(), "prompts")
	legacyDir := filepath.Join(root, "20260308-review-api")
	require.NoError(t, os.MkdirAll(legacyDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(legacyDir, prompt.FileName), []byte("Legacy content"), 0644))

	store := prompt.NewStorage(root)
	items, err := store.ListAll()
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "Default", items[0].Category)
	assert.Equal(t, "20260308-review-api", items[0].Name)
	_, err = os.Stat(filepath.Join(root, "Default", "20260308-review-api", prompt.FileName))
	assert.NoError(t, err)
}

func TestStorageUpdateAllowsCaseOnlyRename(t *testing.T) {
	root := filepath.Join(t.TempDir(), "prompts")
	store := prompt.NewStorage(root)

	_, err := store.Create("Gitacp", "Git helper", "Default", "git add && git commit && git push")
	require.NoError(t, err)

	updated, err := store.Update("Gitacp", "gitacp", "Git helper", "Default", "git add && git commit && git push")
	require.NoError(t, err)
	assert.Equal(t, "gitacp", updated.Name)
	assert.Equal(t, filepath.Join(root, "Default", "gitacp"), updated.Path)

	items, err := store.ListAll()
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "gitacp", items[0].Name)
	assert.Equal(t, filepath.Join(root, "Default", "gitacp"), items[0].Path)

	_, err = os.Stat(filepath.Join(root, "Default", "gitacp", prompt.FileName))
	assert.NoError(t, err)

	entries, err := os.ReadDir(filepath.Join(root, "Default"))
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "gitacp", entries[0].Name())
}
