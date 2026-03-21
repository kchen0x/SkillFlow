package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultSettings(t *testing.T) {
	root := t.TempDir()
	settings := DefaultSettings(root)
	assert.Equal(t, DefaultCategoryName, settings.Shared.DefaultCategory)
	assert.Equal(t, LocalSettings{}, settings.Local)
}

func TestNormalizeLocalSettingsLeavesSettingsUnchanged(t *testing.T) {
	root := t.TempDir()
	normalized := NormalizeLocalSettings(LocalSettings{}, root)
	assert.Equal(t, LocalSettings{}, normalized)
}
