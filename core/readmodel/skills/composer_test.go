package skills

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shinerio/skillflow/core/agentintegration/domain"
	"github.com/shinerio/skillflow/core/shared/logicalkey"
	skillquery "github.com/shinerio/skillflow/core/skillcatalog/app/query"
	skilldomain "github.com/shinerio/skillflow/core/skillcatalog/domain"
	sourcedomain "github.com/shinerio/skillflow/core/skillsource/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildInstalledSkillEntriesUsesDefaultCategoryAndPresence(t *testing.T) {
	logical := logicalkey.Git("github.com/acme/skills", "alpha")
	presence := domain.NewAgentPresenceIndex()
	presence.Add("codex", logical)

	entries := buildInstalledSkillEntries([]*skilldomain.InstalledSkill{
		{
			ID:            "id-1",
			Name:          "alpha",
			Category:      "",
			Source:        skilldomain.SourceGitHub,
			SourceURL:     "https://github.com/acme/skills",
			SourceSubPath: "alpha",
			SourceSHA:     "sha-1",
			LatestSHA:     "sha-2",
		},
	}, presence, "General")

	require.Len(t, entries, 1)
	assert.Equal(t, "General", entries[0].Category)
	assert.True(t, entries[0].Pushed)
	assert.Equal(t, []string{"codex"}, entries[0].PushedAgents)
	assert.True(t, entries[0].Updatable)
}

func TestResolveSourceCandidatesUsesGitLogicalKeyAndContentAliasPresence(t *testing.T) {
	candidateDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(candidateDir, "SKILL.md"), []byte("hello"), 0644))
	contentKey, err := logicalkey.ContentFromDir(candidateDir)
	require.NoError(t, err)

	idx := skillquery.BuildInstalledIndex([]*skilldomain.InstalledSkill{
		{
			ID:            "id-1",
			Name:          "alpha",
			Source:        skilldomain.SourceGitHub,
			SourceURL:     "https://github.com/acme/skills",
			SourceSubPath: "alpha",
			SourceSHA:     "sha-1",
			LatestSHA:     "sha-2",
		},
	})
	presence := domain.NewAgentPresenceIndex()
	presence.Add("claude", contentKey)

	candidates := resolveSourceCandidates([]sourcedomain.SourceSkillCandidate{
		{
			Name:     "alpha",
			Path:     candidateDir,
			SubPath:  "alpha",
			RepoURL:  "https://github.com/acme/skills",
			RepoName: "skills",
			Source:   "github.com/acme/skills",
		},
	}, idx, presence)

	require.Len(t, candidates, 1)
	assert.Equal(t, logicalkey.Git("github.com/acme/skills", "alpha"), candidates[0].LogicalKey)
	assert.True(t, candidates[0].Installed)
	assert.True(t, candidates[0].Imported)
	assert.True(t, candidates[0].Updatable)
	assert.True(t, candidates[0].Pushed)
	assert.Equal(t, []string{"claude"}, candidates[0].PushedAgents)
}
