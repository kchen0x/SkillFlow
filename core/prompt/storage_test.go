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

type exportBundleFixture struct {
	Version int                   `json:"version"`
	Prompts []exportPromptFixture `json:"prompts"`
}

type exportPromptFixture struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Category    string              `json:"category"`
	Content     string              `json:"content"`
	ImageURLs   []string            `json:"imageURLs"`
	WebLinks    []prompt.PromptLink `json:"webLinks"`
}

func TestStorageCreateUpdateMoveDelete(t *testing.T) {
	root := filepath.Join(t.TempDir(), "prompts")
	store := prompt.NewStorage(root)

	created, err := store.Create(
		"Review API",
		"Review backend changes",
		"Default",
		"Please review the API diff.",
		[]string{
			"https://cdn.example.com/review-1.png",
			"https://cdn.example.com/review-2.png",
		},
		[]prompt.PromptLink{{
			Label: "PRD",
			URL:   "https://docs.example.com/prd",
		}},
	)
	require.NoError(t, err)
	assert.Equal(t, "Review API", created.Name)
	assert.Equal(t, "Default", created.Category)
	assert.Equal(t, filepath.Join(root, "Default", "Review API", prompt.FileName), created.FilePath)
	assert.Equal(t, []string{
		"https://cdn.example.com/review-1.png",
		"https://cdn.example.com/review-2.png",
	}, created.ImageURLs)
	assert.Equal(t, []prompt.PromptLink{{
		Label: "PRD",
		URL:   "https://docs.example.com/prd",
	}}, created.WebLinks)

	updated, err := store.Update(
		"Review API",
		"Review API",
		"Review backend and frontend changes",
		"Writing",
		"Please review the UI diff too.",
		[]string{"https://cdn.example.com/review-3.png"},
		[]prompt.PromptLink{{
			Label: "Preview",
			URL:   "https://preview.example.com/review",
		}},
	)
	require.NoError(t, err)
	assert.Equal(t, "Writing", updated.Category)
	assert.Equal(t, "Review backend and frontend changes", updated.Description)
	assert.Equal(t, "Please review the UI diff too.", updated.Content)
	assert.Equal(t, []string{"https://cdn.example.com/review-3.png"}, updated.ImageURLs)
	assert.Equal(t, []prompt.PromptLink{{
		Label: "Preview",
		URL:   "https://preview.example.com/review",
	}}, updated.WebLinks)

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

	_, err = store.Create("Architecture", "", "Research", "Summarize the architecture.", nil, nil)
	require.NoError(t, err)

	err = store.DeleteCategory("Research")
	assert.ErrorIs(t, err, prompt.ErrCategoryNotEmpty)

	require.NoError(t, store.MoveCategory("Architecture", "Default"))
	require.NoError(t, store.DeleteCategory("Research"))
}

func TestStorageExportImportJSON(t *testing.T) {
	root := filepath.Join(t.TempDir(), "prompts")
	store := prompt.NewStorage(root)
	_, err := store.Create("Prompt A", "Desc A", "Default", "Content A", []string{"https://cdn.example.com/prompt-a.png"}, []prompt.PromptLink{{
		Label: "Repo",
		URL:   "https://github.com/example/prompt-a",
	}})
	require.NoError(t, err)
	_, err = store.Create("Prompt B", "Desc B", "Ops", "Content B", nil, nil)
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
	assert.Equal(t, []string{"https://cdn.example.com/prompt-a.png"}, items[0].ImageURLs)
	assert.Equal(t, []prompt.PromptLink{{
		Label: "Repo",
		URL:   "https://github.com/example/prompt-a",
	}}, items[0].WebLinks)
}

func TestStorageExportJSONByNamesReturnsAllWhenEmpty(t *testing.T) {
	root := filepath.Join(t.TempDir(), "prompts")
	store := prompt.NewStorage(root)

	_, err := store.Create("Prompt A", "Desc A", "Default", "Content A", []string{"https://cdn.example.com/a.png"}, []prompt.PromptLink{{
		Label: "Doc A",
		URL:   "https://docs.example.com/a",
	}})
	require.NoError(t, err)
	_, err = store.Create("Prompt B", "Desc B", "Writing", "Content B", nil, []prompt.PromptLink{{
		Label: "Doc B",
		URL:   "https://docs.example.com/b",
	}})
	require.NoError(t, err)

	data, err := store.ExportJSONByNames(nil)
	require.NoError(t, err)

	var bundle exportBundleFixture
	require.NoError(t, json.Unmarshal(data, &bundle))
	require.Len(t, bundle.Prompts, 2)
	assert.ElementsMatch(t, []string{"Prompt A", "Prompt B"}, []string{bundle.Prompts[0].Name, bundle.Prompts[1].Name})
}

func TestStorageExportJSONByNamesFiltersPromptSubset(t *testing.T) {
	root := filepath.Join(t.TempDir(), "prompts")
	store := prompt.NewStorage(root)

	_, err := store.Create("Prompt A", "Desc A", "Default", "Content A", []string{"https://cdn.example.com/a.png"}, []prompt.PromptLink{{
		Label: "Doc A",
		URL:   "https://docs.example.com/a",
	}})
	require.NoError(t, err)
	_, err = store.Create("Prompt B", "Desc B", "Writing", "Content B", []string{"https://cdn.example.com/b.png"}, []prompt.PromptLink{{
		Label: "Doc B",
		URL:   "https://docs.example.com/b",
	}})
	require.NoError(t, err)
	_, err = store.Create("Prompt C", "Desc C", "Research", "Content C", nil, nil)
	require.NoError(t, err)

	data, err := store.ExportJSONByNames([]string{"Prompt B"})
	require.NoError(t, err)

	var bundle exportBundleFixture
	require.NoError(t, json.Unmarshal(data, &bundle))
	require.Len(t, bundle.Prompts, 1)
	assert.Equal(t, exportPromptFixture{
		Name:        "Prompt B",
		Description: "Desc B",
		Category:    "Writing",
		Content:     "Content B",
		ImageURLs:   []string{"https://cdn.example.com/b.png"},
		WebLinks: []prompt.PromptLink{{
			Label: "Doc B",
			URL:   "https://docs.example.com/b",
		}},
	}, bundle.Prompts[0])
}

func TestStorageImportArrayPayloadUpdatesExistingPrompt(t *testing.T) {
	root := filepath.Join(t.TempDir(), "prompts")
	store := prompt.NewStorage(root)
	_, err := store.Create("Prompt A", "Old", "Default", "Old content", nil, nil)
	require.NoError(t, err)

	payload, err := json.Marshal([]map[string]any{{
		"name":        "Prompt A",
		"description": "New",
		"category":    "Research",
		"content":     "New content",
		"imageURLs":   []string{"https://cdn.example.com/new.png"},
		"webLinks": []map[string]string{{
			"label": "Spec",
			"url":   "https://docs.example.com/spec",
		}},
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
	assert.Equal(t, []string{"https://cdn.example.com/new.png"}, item.ImageURLs)
	assert.Equal(t, []prompt.PromptLink{{
		Label: "Spec",
		URL:   "https://docs.example.com/spec",
	}}, item.WebLinks)
}

func TestStoragePreviewImportJSONSeparatesCreatesAndConflicts(t *testing.T) {
	root := filepath.Join(t.TempDir(), "prompts")
	store := prompt.NewStorage(root)

	_, err := store.Create("Prompt A", "Existing", "Default", "Existing content", nil, nil)
	require.NoError(t, err)

	payload, err := json.Marshal(exportBundleFixture{
		Version: 2,
		Prompts: []exportPromptFixture{
			{
				Name:        "Prompt A",
				Description: "Imported existing",
				Category:    "Writing",
				Content:     "Imported content",
			},
			{
				Name:        "Prompt B",
				Description: "Imported new",
				Category:    "Research",
				Content:     "New content",
			},
		},
	})
	require.NoError(t, err)

	preview, err := store.PreviewImportJSON(payload)
	require.NoError(t, err)
	require.Len(t, preview.Creates, 1)
	require.Len(t, preview.Conflicts, 1)
	assert.Equal(t, "Prompt B", preview.Creates[0].Name)
	assert.Equal(t, "Research", preview.Creates[0].Category)
	assert.Equal(t, "Prompt A", preview.Conflicts[0].Name)
	assert.Equal(t, "Writing", preview.Conflicts[0].Category)

	items, err := store.ListAll()
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "Prompt A", items[0].Name)
}

func TestStoragePreviewImportJSONDeduplicatesCreatesWithinImportFile(t *testing.T) {
	root := filepath.Join(t.TempDir(), "prompts")
	store := prompt.NewStorage(root)

	payload, err := json.Marshal(exportBundleFixture{
		Version: 2,
		Prompts: []exportPromptFixture{
			{
				Name:        "Prompt A",
				Description: "First import",
				Category:    "Default",
				Content:     "First content",
			},
			{
				Name:        "Prompt A",
				Description: "Second import",
				Category:    "Writing",
				Content:     "Second content",
			},
		},
	})
	require.NoError(t, err)

	preview, err := store.PreviewImportJSON(payload)
	require.NoError(t, err)
	require.Len(t, preview.Creates, 1)
	require.Empty(t, preview.Conflicts)
	assert.Equal(t, "Prompt A", preview.Creates[0].Name)
	assert.Equal(t, "Second import", preview.Creates[0].Description)
	assert.Equal(t, "Writing", preview.Creates[0].Category)
	assert.Equal(t, "Second content", preview.Creates[0].Content)

	count, err := store.ApplyImportPreview(preview, nil)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	item, err := store.Get("Prompt A")
	require.NoError(t, err)
	assert.Equal(t, "Second import", item.Description)
	assert.Equal(t, "Writing", item.Category)
	assert.Equal(t, "Second content", item.Content)
}

func TestStorageApplyImportSkipsConflicts(t *testing.T) {
	root := filepath.Join(t.TempDir(), "prompts")
	store := prompt.NewStorage(root)

	_, err := store.Create("Prompt A", "Existing", "Default", "Existing content", nil, nil)
	require.NoError(t, err)

	payload, err := json.Marshal(exportBundleFixture{
		Version: 2,
		Prompts: []exportPromptFixture{
			{
				Name:        "Prompt A",
				Description: "Imported existing",
				Category:    "Writing",
				Content:     "Imported content",
			},
			{
				Name:        "Prompt B",
				Description: "Imported new",
				Category:    "Research",
				Content:     "New content",
			},
		},
	})
	require.NoError(t, err)

	preview, err := store.PreviewImportJSON(payload)
	require.NoError(t, err)

	count, err := store.ApplyImportPreview(preview, nil)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	existing, err := store.Get("Prompt A")
	require.NoError(t, err)
	assert.Equal(t, "Default", existing.Category)
	assert.Equal(t, "Existing", existing.Description)
	assert.Equal(t, "Existing content", existing.Content)

	created, err := store.Get("Prompt B")
	require.NoError(t, err)
	assert.Equal(t, "Research", created.Category)
	assert.Equal(t, "Imported new", created.Description)
	assert.Equal(t, "New content", created.Content)
}

func TestStorageApplyImportOverwritesConflictAndCategory(t *testing.T) {
	root := filepath.Join(t.TempDir(), "prompts")
	store := prompt.NewStorage(root)

	_, err := store.Create("Prompt A", "Existing", "Default", "Existing content", []string{"https://cdn.example.com/existing.png"}, []prompt.PromptLink{{
		Label: "Old",
		URL:   "https://docs.example.com/old",
	}})
	require.NoError(t, err)

	payload, err := json.Marshal(exportBundleFixture{
		Version: 2,
		Prompts: []exportPromptFixture{{
			Name:        "Prompt A",
			Description: "Imported existing",
			Category:    "Writing",
			Content:     "Imported content",
			ImageURLs:   []string{"https://cdn.example.com/imported.png"},
			WebLinks: []prompt.PromptLink{{
				Label: "New",
				URL:   "https://docs.example.com/new",
			}},
		}},
	})
	require.NoError(t, err)

	preview, err := store.PreviewImportJSON(payload)
	require.NoError(t, err)

	count, err := store.ApplyImportPreview(preview, []string{"Prompt A"})
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	item, err := store.Get("Prompt A")
	require.NoError(t, err)
	assert.Equal(t, "Writing", item.Category)
	assert.Equal(t, "Imported existing", item.Description)
	assert.Equal(t, "Imported content", item.Content)
	assert.Equal(t, []string{"https://cdn.example.com/imported.png"}, item.ImageURLs)
	assert.Equal(t, []prompt.PromptLink{{
		Label: "New",
		URL:   "https://docs.example.com/new",
	}}, item.WebLinks)
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

	_, err := store.Create("Gitacp", "Git helper", "Default", "git add && git commit && git push", nil, nil)
	require.NoError(t, err)

	updated, err := store.Update("Gitacp", "gitacp", "Git helper", "Default", "git add && git commit && git push", nil, nil)
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

func TestStorageCreateRejectsTooManyImages(t *testing.T) {
	root := filepath.Join(t.TempDir(), "prompts")
	store := prompt.NewStorage(root)

	_, err := store.Create(
		"Review API",
		"Review backend changes",
		"Default",
		"Please review the API diff.",
		[]string{
			"https://cdn.example.com/1.png",
			"https://cdn.example.com/2.png",
			"https://cdn.example.com/3.png",
			"https://cdn.example.com/4.png",
		},
		nil,
	)
	assert.ErrorIs(t, err, prompt.ErrTooManyImages)
}
