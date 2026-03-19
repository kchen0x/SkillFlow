package skillkey_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/skillkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitFromRepoURLNormalizesRemoteAndSubPath(t *testing.T) {
	key, err := skillkey.GitFromRepoURL("git@github.com:OpenAI/Skills.git", `skills\my-skill`)
	require.NoError(t, err)
	assert.Equal(t, "git:github.com/openai/skills#skills/my-skill", key)
}

func TestGitFromRepoURLSupportsRepoRootSkill(t *testing.T) {
	key, err := skillkey.GitFromRepoURL("https://github.com/OpenAI/Skills", ".")
	require.NoError(t, err)
	assert.Equal(t, "git:github.com/openai/skills#.", key)
}

func TestContentFromDirIsStableAcrossLocations(t *testing.T) {
	left := t.TempDir()
	right := t.TempDir()

	makeSkillDir := func(root string) string {
		dir := filepath.Join(root, "alpha")
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "docs"), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "skill.md"), []byte("# alpha\n"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "docs", "notes.txt"), []byte("same\n"), 0644))
		return dir
	}

	leftDir := makeSkillDir(left)
	rightDir := makeSkillDir(right)

	leftKey, err := skillkey.ContentFromDir(leftDir)
	require.NoError(t, err)
	rightKey, err := skillkey.ContentFromDir(rightDir)
	require.NoError(t, err)

	assert.Equal(t, leftKey, rightKey)
	assert.Contains(t, leftKey, "content:")
}

func TestContentFromDirChangesWhenContentChanges(t *testing.T) {
	root := t.TempDir()
	first := filepath.Join(root, "first")
	second := filepath.Join(root, "second")

	require.NoError(t, os.MkdirAll(first, 0755))
	require.NoError(t, os.MkdirAll(second, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(first, "skill.md"), []byte("# alpha\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(second, "skill.md"), []byte("# beta\n"), 0644))

	firstKey, err := skillkey.ContentFromDir(first)
	require.NoError(t, err)
	secondKey, err := skillkey.ContentFromDir(second)
	require.NoError(t, err)

	assert.NotEqual(t, firstKey, secondKey)
}
