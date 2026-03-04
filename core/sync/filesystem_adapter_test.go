package sync_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/skill"
	toolsync "github.com/shinerio/skillflow/core/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeSkillDir(t *testing.T, root, category, name string) *skill.Skill {
	t.Helper()
	dir := filepath.Join(root, category, name)
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILLS.md"), []byte("# "+name), 0644))
	return &skill.Skill{Name: name, Path: dir, Category: category}
}

func TestFilesystemAdapterPushFlattens(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	sk := makeSkillDir(t, src, "coding", "my-skill")

	adapter := toolsync.NewFilesystemAdapter("test-tool", "")
	err := adapter.Push(context.Background(), []*skill.Skill{sk}, dst)
	require.NoError(t, err)

	// skill should be at dst/my-skill (no category subdir)
	_, err = os.Stat(filepath.Join(dst, "my-skill", "SKILLS.md"))
	assert.NoError(t, err)
}

func TestFilesystemAdapterPull(t *testing.T) {
	src := t.TempDir()
	// Create two valid skills directly in src (tool dir is flat)
	for _, name := range []string{"skill-x", "skill-y"} {
		dir := filepath.Join(src, name)
		require.NoError(t, os.MkdirAll(dir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILLS.md"), []byte("# "+name), 0644))
	}
	// Create a non-skill directory
	require.NoError(t, os.MkdirAll(filepath.Join(src, "not-a-skill"), 0755))

	adapter := toolsync.NewFilesystemAdapter("test-tool", "")
	skills, err := adapter.Pull(context.Background(), src)
	require.NoError(t, err)
	assert.Len(t, skills, 2)
}
