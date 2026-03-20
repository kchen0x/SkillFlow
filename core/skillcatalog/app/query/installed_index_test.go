package query_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/skillcatalog/app/query"
	"github.com/shinerio/skillflow/core/skillcatalog/domain"
	"github.com/shinerio/skillflow/core/skillkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildInstalledIndexResolvesByLogicalKeyBeforeName(t *testing.T) {
	root := t.TempDir()
	makeDir := func(name, content string) string {
		dir := filepath.Join(root, name)
		require.NoError(t, os.MkdirAll(dir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "skill.md"), []byte(content), 0644))
		return dir
	}

	firstPath := makeDir("same-name-a", "# alpha\n")
	secondPath := makeDir("same-name-b", "# beta\n")

	skills := []*domain.InstalledSkill{
		{
			ID:            "skill-a",
			Name:          "same-name",
			Path:          firstPath,
			Source:        domain.SourceGitHub,
			SourceURL:     "https://github.com/acme/repo-a",
			SourceSubPath: "skills/same-name",
		},
		{
			ID:            "skill-b",
			Name:          "same-name",
			Path:          secondPath,
			Source:        domain.SourceGitHub,
			SourceURL:     "https://github.com/acme/repo-b",
			SourceSubPath: "skills/same-name",
		},
	}

	idx := query.BuildInstalledIndex(skills)
	status := idx.Resolve("same-name", "git:github.com/acme/repo-b#skills/same-name")

	assert.True(t, status.Installed)
	assert.Equal(t, skillkey.MatchStrengthLogical, status.MatchStrength)
	assert.Equal(t, "git:github.com/acme/repo-b#skills/same-name", status.LogicalKey)
}

func TestBuildInstalledIndexFallsBackToUniqueNameWhenLogicalKeyMissing(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "manual")
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "skill.md"), []byte("# alpha\n"), 0644))

	idx := query.BuildInstalledIndex([]*domain.InstalledSkill{{
		ID:        "manual-1",
		Name:      "alpha",
		Path:      dir,
		Source:    domain.SourceManual,
		LatestSHA: "",
	}})

	status := idx.Resolve("alpha", "")
	assert.True(t, status.Installed)
	assert.True(t, status.Imported)
	assert.Equal(t, skillkey.MatchStrengthFallback, status.MatchStrength)
}

func TestBuildInstalledIndexMarksGroupUpdatable(t *testing.T) {
	idx := query.BuildInstalledIndex([]*domain.InstalledSkill{{
		ID:            "git-1",
		Name:          "alpha",
		Source:        domain.SourceGitHub,
		SourceURL:     "https://github.com/acme/repo",
		SourceSubPath: "skills/alpha",
		SourceSHA:     "old",
		LatestSHA:     "new",
	}})

	status := idx.Resolve("alpha", "git:github.com/acme/repo#skills/alpha")
	assert.True(t, status.Updatable)
}

func TestBuildInstalledIndexIncludesInstalledContentKeys(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "alpha")
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "skill.md"), []byte("# alpha\nLine 2\n"), 0644))

	idx := query.BuildInstalledIndex([]*domain.InstalledSkill{{
		ID:            "git-1",
		Name:          "alpha",
		Path:          dir,
		Source:        domain.SourceGitHub,
		SourceURL:     "https://github.com/acme/repo",
		SourceSubPath: "skills/alpha",
	}})

	contentKey, err := skillkey.ContentFromDir(dir)
	require.NoError(t, err)

	status := idx.Resolve("alpha", "git:github.com/acme/repo#skills/alpha")
	assert.Contains(t, status.ContentKeys, contentKey)
}

func TestBuildInstalledIndexResolvesGitSkillByContentKey(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "alpha")
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "skill.md"), []byte("# alpha\nLine 2\n"), 0644))

	idx := query.BuildInstalledIndex([]*domain.InstalledSkill{{
		ID:            "git-1",
		Name:          "alpha",
		Path:          dir,
		Source:        domain.SourceGitHub,
		SourceURL:     "https://github.com/acme/repo",
		SourceSubPath: "skills/alpha",
	}})

	contentKey, err := skillkey.ContentFromDir(dir)
	require.NoError(t, err)

	status := idx.Resolve("alpha", contentKey)
	assert.True(t, status.Installed)
	assert.Equal(t, skillkey.MatchStrengthContent, status.MatchStrength)
	assert.Equal(t, "git:github.com/acme/repo#skills/alpha", status.LogicalKey)
}
