package app

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultSettings(t *testing.T) {
	root := t.TempDir()
	settings := DefaultSettings(root)
	assert.Equal(t, DefaultCategoryName, settings.Shared.DefaultCategory)
	assert.Equal(t, filepath.Join(root, "skills"), settings.Local.SkillsStorageDir)
}

func TestNormalizeLocalSettingsUsesDefaultSkillsDir(t *testing.T) {
	root := t.TempDir()
	normalized := NormalizeLocalSettings(LocalSettings{}, root)
	assert.Equal(t, filepath.Join(root, "skills"), normalized.SkillsStorageDir)
}
