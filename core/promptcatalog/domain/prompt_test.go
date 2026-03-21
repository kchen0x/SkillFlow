package domain_test

import (
	"testing"

	"github.com/shinerio/skillflow/core/promptcatalog/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseWebLinksMarkdown(t *testing.T) {
	links, err := domain.ParseWebLinksMarkdown("[Doc](https://docs.example.com/a)\n[Preview](https://preview.example.com/b)")
	require.NoError(t, err)
	assert.Equal(t, []domain.PromptLink{
		{Label: "Doc", URL: "https://docs.example.com/a"},
		{Label: "Preview", URL: "https://preview.example.com/b"},
	}, links)
}

func TestParseWebLinksMarkdownRejectsInvalidFormat(t *testing.T) {
	_, err := domain.ParseWebLinksMarkdown("https://docs.example.com/a")
	assert.ErrorIs(t, err, domain.ErrInvalidWebLink)
}

func TestNormalizePromptImageURLsRejectsTooManyImages(t *testing.T) {
	_, err := domain.NormalizePromptImageURLs([]string{
		"https://cdn.example.com/1.png",
		"https://cdn.example.com/2.png",
		"https://cdn.example.com/3.png",
		"https://cdn.example.com/4.png",
	})
	assert.ErrorIs(t, err, domain.ErrTooManyImages)
}

func TestNormalizeCategoryNameDefaultsToDefault(t *testing.T) {
	category, err := domain.NormalizeCategoryName("")
	require.NoError(t, err)
	assert.Equal(t, domain.DefaultCategoryName, category)
}
