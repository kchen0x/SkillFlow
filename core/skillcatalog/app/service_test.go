package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/skillcatalog/app"
	"github.com/shinerio/skillflow/core/skillcatalog/domain"
	repositoryinfra "github.com/shinerio/skillflow/core/skillcatalog/infra/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceImportAndListAll(t *testing.T) {
	root := filepath.Join(t.TempDir(), "skills")
	require.NoError(t, os.MkdirAll(root, 0755))
	srcRoot := t.TempDir()
	srcDir := filepath.Join(srcRoot, "demo-skill")
	require.NoError(t, os.MkdirAll(srcDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "skill.md"), []byte("# demo"), 0644))

	svc := app.NewService(repositoryinfra.NewFilesystemStorage(root))
	imported, err := svc.Import(srcDir, "coding", domain.SourceManual, "", "")
	require.NoError(t, err)

	items, err := svc.ListAll()
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, imported.ID, items[0].ID)
	assert.Equal(t, "demo-skill", items[0].Name)
}
