package app_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/promptcatalog/app"
	"github.com/shinerio/skillflow/core/promptcatalog/domain"
	repositoryinfra "github.com/shinerio/skillflow/core/promptcatalog/infra/repository"
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
	WebLinks    []domain.PromptLink `json:"webLinks"`
}

func TestServiceCreateUpdateMoveDelete(t *testing.T) {
	root := filepath.Join(t.TempDir(), "prompts")
	svc := app.NewService(repositoryinfra.NewFilesystemStorage(root))

	created, err := svc.CreatePrompt(
		"Review API",
		"Review backend changes",
		"Default",
		"Please review the API diff.",
		[]string{
			"https://cdn.example.com/review-1.png",
			"https://cdn.example.com/review-2.png",
		},
		[]domain.PromptLink{{
			Label: "PRD",
			URL:   "https://docs.example.com/prd",
		}},
	)
	require.NoError(t, err)
	assert.Equal(t, "Review API", created.Name)
	assert.Equal(t, "Default", created.Category)
	assert.Equal(t, filepath.Join(root, "Default", "Review API", domain.FileName), created.FilePath)

	updated, err := svc.UpdatePrompt(
		"Review API",
		"Review API",
		"Review backend and frontend changes",
		"Writing",
		"Please review the UI diff too.",
		[]string{"https://cdn.example.com/review-3.png"},
		[]domain.PromptLink{{
			Label: "Preview",
			URL:   "https://preview.example.com/review",
		}},
	)
	require.NoError(t, err)
	assert.Equal(t, "Writing", updated.Category)
	assert.Equal(t, "Please review the UI diff too.", updated.Content)

	items, err := svc.ListPrompts()
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "Writing", items[0].Category)

	require.NoError(t, svc.DeletePrompt("Review API"))
	_, err = os.Stat(filepath.Join(root, "Writing", "Review API"))
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestServiceCategories(t *testing.T) {
	root := filepath.Join(t.TempDir(), "prompts")
	svc := app.NewService(repositoryinfra.NewFilesystemStorage(root))
	require.NoError(t, svc.CreatePromptCategory("Default"))
	require.NoError(t, svc.CreatePromptCategory("Research"))

	categories, err := svc.ListPromptCategories()
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"Default", "Research"}, categories)

	_, err = svc.CreatePrompt("Architecture", "", "Research", "Summarize the architecture.", nil, nil)
	require.NoError(t, err)

	err = svc.DeletePromptCategory("Research")
	assert.ErrorIs(t, err, domain.ErrCategoryNotEmpty)

	require.NoError(t, svc.MovePromptToCategory("Architecture", "Default"))
	require.NoError(t, svc.DeletePromptCategory("Research"))
}

func TestServiceExportImportJSON(t *testing.T) {
	root := filepath.Join(t.TempDir(), "prompts")
	svc := app.NewService(repositoryinfra.NewFilesystemStorage(root))
	_, err := svc.CreatePrompt("Prompt A", "Desc A", "Default", "Content A", []string{"https://cdn.example.com/prompt-a.png"}, []domain.PromptLink{{
		Label: "Repo",
		URL:   "https://github.com/example/prompt-a",
	}})
	require.NoError(t, err)
	_, err = svc.CreatePrompt("Prompt B", "Desc B", "Ops", "Content B", nil, nil)
	require.NoError(t, err)

	data, err := svc.ExportPromptBundle(nil)
	require.NoError(t, err)
	assert.Contains(t, string(data), "Prompt A")
	assert.Contains(t, string(data), "Prompt B")

	other := app.NewService(repositoryinfra.NewFilesystemStorage(filepath.Join(t.TempDir(), "prompts")))
	preview, err := other.PreviewPromptImport(data)
	require.NoError(t, err)
	count, err := other.ApplyPromptImport(preview, []string{"Prompt A", "Prompt B"})
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	items, err := other.ListPrompts()
	require.NoError(t, err)
	require.Len(t, items, 2)
}

func TestServicePreviewImportSeparatesCreatesAndConflicts(t *testing.T) {
	root := filepath.Join(t.TempDir(), "prompts")
	svc := app.NewService(repositoryinfra.NewFilesystemStorage(root))

	_, err := svc.CreatePrompt("Prompt A", "Existing", "Default", "Existing content", nil, nil)
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

	preview, err := svc.PreviewPromptImport(payload)
	require.NoError(t, err)
	require.Len(t, preview.Creates, 1)
	require.Len(t, preview.Conflicts, 1)
	assert.Equal(t, "Prompt B", preview.Creates[0].Name)
	assert.Equal(t, "Prompt A", preview.Conflicts[0].Name)
}
