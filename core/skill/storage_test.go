package skill_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/shinerio/skillflow/core/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeTestSkillDir(t *testing.T, baseDir, name string) string {
	t.Helper()
	dir := filepath.Join(baseDir, name)
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "skill.md"), []byte("# "+name), 0644))
	return dir
}

func writeStoredMeta(t *testing.T, root string, sk skill.Skill) {
	t.Helper()
	metaDir := filepath.Join(filepath.Dir(root), "meta")
	require.NoError(t, os.MkdirAll(metaDir, 0755))
	data, err := json.MarshalIndent(sk, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(metaDir, sk.ID+".json"), data, 0644))
}

func readStoredPath(t *testing.T, root, id string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(filepath.Dir(root), "meta", id+".json"))
	require.NoError(t, err)
	var stored skill.Skill
	require.NoError(t, json.Unmarshal(data, &stored))
	return stored.Path
}

func readStoredMetaRaw(t *testing.T, root, id string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(filepath.Dir(root), "meta", id+".json"))
	require.NoError(t, err)
	return string(data)
}

func foreignAbsolutePath() string {
	if runtime.GOOS == "windows" {
		return "/Users/demo/.skillflow/skills/coding/portable-skill"
	}
	return `C:\Users\demo\.skillflow\skills\coding\portable-skill`
}

func TestStorageListCategories(t *testing.T) {
	root := t.TempDir()
	svc := skill.NewStorage(root)
	require.NoError(t, svc.CreateCategory("coding"))
	require.NoError(t, svc.CreateCategory("writing"))
	cats, err := svc.ListCategories()
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"coding", "writing"}, cats)
}

func TestStorageImportSkill(t *testing.T) {
	root := filepath.Join(t.TempDir(), "skills")
	require.NoError(t, os.MkdirAll(root, 0755))
	src := t.TempDir()
	skillDir := makeTestSkillDir(t, src, "my-skill")
	svc := skill.NewStorage(root)

	imported, err := svc.Import(skillDir, "coding", skill.SourceManual, "", "")
	require.NoError(t, err)
	assert.Equal(t, "my-skill", imported.Name)
	assert.Equal(t, "coding", imported.Category)

	// verify directory was copied
	_, err = os.Stat(filepath.Join(root, "coding", "my-skill", "skill.md"))
	assert.NoError(t, err)
	assert.Equal(t, "skills/coding/my-skill", readStoredPath(t, root, imported.ID))
}

func TestStorageConflictDetected(t *testing.T) {
	root := t.TempDir()
	src := t.TempDir()
	skillDir := makeTestSkillDir(t, src, "dup-skill")
	svc := skill.NewStorage(root)

	_, err := svc.Import(skillDir, "coding", skill.SourceManual, "", "")
	require.NoError(t, err)

	_, err = svc.Import(skillDir, "coding", skill.SourceManual, "", "")
	assert.ErrorIs(t, err, skill.ErrSkillExists)
}

func TestStorageDeleteSkill(t *testing.T) {
	root := t.TempDir()
	src := t.TempDir()
	skillDir := makeTestSkillDir(t, src, "del-skill")
	svc := skill.NewStorage(root)

	s, err := svc.Import(skillDir, "", skill.SourceManual, "", "")
	require.NoError(t, err)
	require.NoError(t, svc.Delete(s.ID))

	skills, err := svc.ListAll()
	require.NoError(t, err)
	assert.Empty(t, skills)
}

func TestStorageMoveCategory(t *testing.T) {
	root := t.TempDir()
	src := t.TempDir()
	skillDir := makeTestSkillDir(t, src, "move-skill")
	svc := skill.NewStorage(root)
	require.NoError(t, svc.CreateCategory("cat-a"))
	require.NoError(t, svc.CreateCategory("cat-b"))

	s, err := svc.Import(skillDir, "cat-a", skill.SourceManual, "", "")
	require.NoError(t, err)

	err = svc.MoveCategory(s.ID, "cat-b")
	require.NoError(t, err)

	updated, err := svc.Get(s.ID)
	require.NoError(t, err)
	assert.Equal(t, "cat-b", updated.Category)
}

func TestStorageDeleteEmptyCategory(t *testing.T) {
	root := t.TempDir()
	svc := skill.NewStorage(root)
	require.NoError(t, svc.CreateCategory("empty-cat"))

	require.NoError(t, svc.DeleteCategory("empty-cat"))
	_, err := os.Stat(filepath.Join(root, "empty-cat"))
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestStorageDeleteCategoryRejectsWhenNotEmpty(t *testing.T) {
	root := t.TempDir()
	src := t.TempDir()
	skillDir := makeTestSkillDir(t, src, "busy-skill")
	svc := skill.NewStorage(root)
	require.NoError(t, svc.CreateCategory("busy-cat"))
	_, err := svc.Import(skillDir, "busy-cat", skill.SourceManual, "", "")
	require.NoError(t, err)

	err = svc.DeleteCategory("busy-cat")
	assert.ErrorIs(t, err, skill.ErrCategoryNotEmpty)
	_, statErr := os.Stat(filepath.Join(root, "busy-cat"))
	assert.NoError(t, statErr)
}

func TestStorageListAllMigratesAbsoluteMetaPath(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "skills")
	actual := filepath.Join(root, "coding", "portable-skill")
	require.NoError(t, os.MkdirAll(actual, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(actual, "skill.md"), []byte("# portable-skill"), 0644))

	stored := skill.Skill{
		ID:       "skill-absolute",
		Name:     "portable-skill",
		Category: "coding",
		Path:     actual,
	}
	writeStoredMeta(t, root, stored)

	svc := skill.NewStorage(root)
	items, err := svc.ListAll()
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, actual, items[0].Path)
	assert.Equal(t, "skills/coding/portable-skill", readStoredPath(t, root, stored.ID))
}

func TestStorageListAllRecoversFromForeignAbsoluteMetaPath(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "skills")
	actual := filepath.Join(root, "coding", "portable-skill")
	require.NoError(t, os.MkdirAll(actual, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(actual, "skill.md"), []byte("# portable-skill"), 0644))

	stored := skill.Skill{
		ID:       "skill-foreign",
		Name:     "portable-skill",
		Category: "coding",
		Path:     foreignAbsolutePath(),
	}
	writeStoredMeta(t, root, stored)

	svc := skill.NewStorage(root)
	items, err := svc.ListAll()
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, actual, items[0].Path)
	assert.Equal(t, "skills/coding/portable-skill", readStoredPath(t, root, stored.ID))
}

func TestStorageSaveMetaStoresLastCheckedAtInLocalMetaOnly(t *testing.T) {
	root := filepath.Join(t.TempDir(), "skills")
	require.NoError(t, os.MkdirAll(root, 0755))
	src := t.TempDir()
	skillDir := makeTestSkillDir(t, src, "local-meta-skill")
	svc := skill.NewStorage(root)

	imported, err := svc.Import(skillDir, "coding", skill.SourceGitHub, "https://github.com/example/repo.git", "skills/local-meta-skill")
	require.NoError(t, err)
	checkedAt := time.Now().UTC().Truncate(time.Second)
	imported.LastCheckedAt = checkedAt
	require.NoError(t, svc.SaveMeta(imported))

	metaRaw := readStoredMetaRaw(t, root, imported.ID)
	if strings.Contains(metaRaw, "LastCheckedAt") {
		t.Fatalf("expected synced meta to exclude LastCheckedAt, got %s", metaRaw)
	}

	localPath := filepath.Join(filepath.Dir(root), "meta_local", imported.ID+".local.json")
	localRaw, err := os.ReadFile(localPath)
	require.NoError(t, err)
	assert.Contains(t, string(localRaw), "\"lastCheckedAt\"")

	reloaded, err := svc.Get(imported.ID)
	require.NoError(t, err)
	assert.Equal(t, checkedAt, reloaded.LastCheckedAt.UTC().Truncate(time.Second))
}

func TestStorageDeleteRemovesLocalMeta(t *testing.T) {
	root := filepath.Join(t.TempDir(), "skills")
	require.NoError(t, os.MkdirAll(root, 0755))
	src := t.TempDir()
	skillDir := makeTestSkillDir(t, src, "delete-local-meta-skill")
	svc := skill.NewStorage(root)

	imported, err := svc.Import(skillDir, "coding", skill.SourceGitHub, "https://github.com/example/repo.git", "skills/delete-local-meta-skill")
	require.NoError(t, err)
	imported.LastCheckedAt = time.Now().UTC()
	require.NoError(t, svc.SaveMeta(imported))

	localPath := filepath.Join(filepath.Dir(root), "meta_local", imported.ID+".local.json")
	_, err = os.Stat(localPath)
	require.NoError(t, err)

	require.NoError(t, svc.Delete(imported.ID))

	_, err = os.Stat(localPath)
	assert.ErrorIs(t, err, os.ErrNotExist)
}
