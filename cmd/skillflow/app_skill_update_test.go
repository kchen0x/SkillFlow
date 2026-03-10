package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/config"
	"github.com/shinerio/skillflow/core/install"
	"github.com/shinerio/skillflow/core/skill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateSkillRefreshesExistingPushedCopies(t *testing.T) {
	app, codexPushDir, claudePushDir := newUpdateSkillTestApp(t)
	sourceDir := writeTestSkillDir(t, t.TempDir(), "demo-skill", "# Demo\nOld\n")

	sk, err := app.storage.Import(sourceDir, defaultCategoryName, skill.SourceGitHub, "https://github.com/octo/demo", "skills/demo-skill")
	require.NoError(t, err)
	sk.SourceSHA = "oldsha"
	sk.LatestSHA = "newsha"
	require.NoError(t, app.storage.UpdateMeta(sk))

	app.autoPushImportedSkills("test.setup", []*skill.Skill{sk})
	conflicts, err := app.PushToTools([]string{sk.ID}, []string{"claude-code"})
	require.NoError(t, err)
	require.Empty(t, conflicts)

	codexSkillPath := filepath.Join(codexPushDir, "demo-skill", "skill.md")
	claudeSkillPath := filepath.Join(claudePushDir, "demo-skill", "skill.md")
	assertFileContentEquals(t, filepath.Join(sk.Path, "skill.md"), "# Demo\nOld\n")
	assertFileContentEquals(t, codexSkillPath, "# Demo\nOld\n")
	assertFileContentEquals(t, claudeSkillPath, "# Demo\nOld\n")

	app.ghDownloader = func(_ *http.Client) githubSkillDownloader {
		return fakeGitHubDownloader{content: "# Demo\nNew\n"}
	}

	require.NoError(t, app.UpdateSkill(sk.ID))

	assertFileContentEquals(t, filepath.Join(sk.Path, "skill.md"), "# Demo\nNew\n")
	assertFileContentEquals(t, codexSkillPath, "# Demo\nNew\n")
	assertFileContentEquals(t, claudeSkillPath, "# Demo\nNew\n")

	updated, err := app.storage.Get(sk.ID)
	require.NoError(t, err)
	assert.Equal(t, "newsha", updated.SourceSHA)
	assert.Empty(t, updated.LatestSHA)
}

func newUpdateSkillTestApp(t *testing.T) (*App, string, string) {
	t.Helper()

	dataDir := t.TempDir()
	codexPushDir := filepath.Join(dataDir, "codex-skills")
	claudePushDir := filepath.Join(dataDir, "claude-skills")
	skillsDir := filepath.Join(dataDir, "library", "skills")

	svc := config.NewService(dataDir)
	cfg := config.DefaultConfig(dataDir)
	cfg.SkillsStorageDir = skillsDir
	cfg.AutoPushTools = []string{"codex"}
	cfg.Tools = []config.ToolConfig{
		{
			Name:     "codex",
			ScanDirs: []string{codexPushDir},
			PushDir:  codexPushDir,
			Enabled:  true,
		},
		{
			Name:     "claude-code",
			ScanDirs: []string{claudePushDir},
			PushDir:  claudePushDir,
			Enabled:  true,
		},
	}
	require.NoError(t, svc.Save(cfg))

	app := NewApp()
	app.config = svc
	app.storage = skill.NewStorage(skillsDir)
	return app, codexPushDir, claudePushDir
}

type fakeGitHubDownloader struct {
	content string
}

func (f fakeGitHubDownloader) DownloadTo(_ context.Context, _ install.InstallSource, _ install.SkillCandidate, targetDir string) error {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(targetDir, "skill.md"), []byte(f.content), 0644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(targetDir, "notes.txt"), []byte("updated"), 0644)
}

func assertFileContentEquals(t *testing.T, path string, want string) {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, want, string(data))
}
