package appdata_test

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/shinerio/skillflow/core/platform/appdata"
	"github.com/stretchr/testify/assert"
)

func TestDirUsesExpectedPlatformSuffix(t *testing.T) {
	dir := filepath.Clean(appdata.Dir())

	switch runtime.GOOS {
	case "windows":
		assert.Equal(t, ".skillflow", filepath.Base(dir))
	default:
		expectedSuffix := filepath.Join("Library", "Application Support", "SkillFlow")
		assert.True(t, strings.HasSuffix(dir, expectedSuffix), "dir=%s", dir)
	}
}

func TestSkillsDirJoinsUnderDataDir(t *testing.T) {
	root := filepath.Join("tmp", "skillflow-data")
	assert.Equal(t, filepath.Join(root, "skills"), appdata.SkillsDir(root))
}

func TestRepoCacheDirJoinsUnderDataDir(t *testing.T) {
	root := filepath.Join("tmp", "skillflow-data")
	assert.Equal(t, filepath.Join(root, "cache", "repos"), appdata.RepoCacheDir(root))
}
