package domain_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/skillcatalog/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mkSkillDir(t *testing.T, filename string) string {
	t.Helper()
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "my-skill")
	require.NoError(t, os.MkdirAll(skillDir, 0755))
	if filename != "" {
		require.NoError(t, os.WriteFile(filepath.Join(skillDir, filename), []byte("# skill"), 0644))
	}
	return skillDir
}

func TestValidatorAcceptsSkillMdVariants(t *testing.T) {
	for _, name := range []string{"skill.md", "SKILL.MD", "Skill.md"} {
		t.Run(name, func(t *testing.T) {
			v := domain.NewValidator()
			assert.NoError(t, v.Validate(mkSkillDir(t, name)))
		})
	}
}

func TestValidatorRejectsDirectoryWithoutSkillMd(t *testing.T) {
	v := domain.NewValidator()
	err := v.Validate(mkSkillDir(t, ""))
	assert.ErrorIs(t, err, domain.ErrNoSKILLSmd)
}

func TestValidatorRejectsNonDirectory(t *testing.T) {
	v := domain.NewValidator()
	err := v.Validate("/nonexistent/path")
	assert.Error(t, err)
}
