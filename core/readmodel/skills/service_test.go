package skills

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
	"github.com/shinerio/skillflow/core/readmodel/viewstate"
	"github.com/shinerio/skillflow/core/shared/logicalkey"
	skillquery "github.com/shinerio/skillflow/core/skillcatalog/app/query"
	skilldomain "github.com/shinerio/skillflow/core/skillcatalog/domain"
	sourcedomain "github.com/shinerio/skillflow/core/skillsource/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeInstalledProvider struct {
	skills []*skilldomain.InstalledSkill
	err    error
	calls  int
}

func (f *fakeInstalledProvider) ListAll() ([]*skilldomain.InstalledSkill, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return f.skills, nil
}

type fakeSourceProvider struct {
	all       []sourcedomain.SourceSkillCandidate
	byRepo    map[string][]sourcedomain.SourceSkillCandidate
	errAll    error
	errByRepo error
	allCalls  int
	repoCalls map[string]int
}

func (f *fakeSourceProvider) ListAllSourceCandidates(_ int) ([]sourcedomain.SourceSkillCandidate, error) {
	f.allCalls++
	if f.errAll != nil {
		return nil, f.errAll
	}
	return f.all, nil
}

func (f *fakeSourceProvider) ListSourceCandidatesByRepo(repoURL string, _ int) ([]sourcedomain.SourceSkillCandidate, error) {
	if f.repoCalls == nil {
		f.repoCalls = map[string]int{}
	}
	f.repoCalls[repoURL]++
	if f.errByRepo != nil {
		return nil, f.errByRepo
	}
	return f.byRepo[repoURL], nil
}

type fakePresenceProvider struct {
	presence *agentdomain.AgentPresenceIndex
	err      error
	calls    int
}

func (f *fakePresenceProvider) Resolve(_ context.Context, _ *skillquery.InstalledIndex, _ int, _ []agentdomain.AgentProfile) (*agentdomain.AgentPresenceIndex, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return f.presence, nil
}

func TestServiceListInstalledSkillsUsesSnapshotCache(t *testing.T) {
	skills := []*skilldomain.InstalledSkill{
		{
			ID:            "id-1",
			Name:          "alpha",
			Source:        skilldomain.SourceGitHub,
			SourceURL:     "https://github.com/acme/skills",
			SourceSubPath: "alpha",
			SourceSHA:     "sha-1",
			LatestSHA:     "sha-2",
		},
	}
	logical, err := skilldomain.LogicalKey(skills[0])
	require.NoError(t, err)

	presence := agentdomain.NewAgentPresenceIndex()
	presence.Add("codex", logical)

	installed := &fakeInstalledProvider{skills: skills}
	sources := &fakeSourceProvider{}
	presenceProvider := &fakePresenceProvider{presence: presence}
	cache := viewstate.NewManager(t.TempDir())

	svc := NewService(installed, sources, presenceProvider, cache)
	input := InstalledSkillsInput{
		DefaultCategory:     "General",
		RepoScanMaxDepth:    2,
		SnapshotFingerprint: "fp-installed",
	}

	first, err := svc.ListInstalledSkills(context.Background(), input)
	require.NoError(t, err)
	require.Len(t, first, 1)
	assert.True(t, first[0].Pushed)
	assert.Equal(t, "General", first[0].Category)

	second, err := svc.ListInstalledSkills(context.Background(), input)
	require.NoError(t, err)
	require.Len(t, second, 1)

	assert.Equal(t, 1, installed.calls)
	assert.Equal(t, 1, presenceProvider.calls)
}

func TestServiceListAllStarSkillsUsesSnapshotCache(t *testing.T) {
	candidateDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(candidateDir, "SKILL.md"), []byte("hello"), 0644))
	contentKey, err := logicalkey.ContentFromDir(candidateDir)
	require.NoError(t, err)

	installed := &fakeInstalledProvider{
		skills: []*skilldomain.InstalledSkill{
			{
				ID:            "id-1",
				Name:          "alpha",
				Source:        skilldomain.SourceGitHub,
				SourceURL:     "https://github.com/acme/skills",
				SourceSubPath: "alpha",
			},
		},
	}
	sources := &fakeSourceProvider{
		all: []sourcedomain.SourceSkillCandidate{
			{
				Name:     "alpha",
				Path:     candidateDir,
				SubPath:  "alpha",
				RepoURL:  "https://github.com/acme/skills",
				RepoName: "skills",
				Source:   "github.com/acme/skills",
			},
		},
	}
	presence := agentdomain.NewAgentPresenceIndex()
	presence.Add("claude", contentKey)
	presenceProvider := &fakePresenceProvider{presence: presence}

	svc := NewService(installed, sources, presenceProvider, viewstate.NewManager(t.TempDir()))
	input := StarSkillsInput{
		RepoScanMaxDepth:    2,
		SnapshotFingerprint: "fp-stars",
	}

	first, err := svc.ListAllStarSkills(context.Background(), input)
	require.NoError(t, err)
	require.Len(t, first, 1)
	assert.True(t, first[0].Installed)
	assert.True(t, first[0].Pushed)

	second, err := svc.ListAllStarSkills(context.Background(), input)
	require.NoError(t, err)
	require.Len(t, second, 1)

	assert.Equal(t, 1, sources.allCalls)
	assert.Equal(t, 1, installed.calls)
	assert.Equal(t, 1, presenceProvider.calls)
}

func TestServiceListRepoStarSkillsUsesRepoComposition(t *testing.T) {
	installed := &fakeInstalledProvider{
		skills: []*skilldomain.InstalledSkill{
			{
				ID:            "id-1",
				Name:          "alpha",
				Source:        skilldomain.SourceGitHub,
				SourceURL:     "https://github.com/acme/skills",
				SourceSubPath: "alpha",
			},
		},
	}
	sources := &fakeSourceProvider{
		byRepo: map[string][]sourcedomain.SourceSkillCandidate{
			"https://github.com/acme/skills": {
				{
					Name:     "alpha",
					SubPath:  "alpha",
					RepoURL:  "https://github.com/acme/skills",
					RepoName: "skills",
					Source:   "github.com/acme/skills",
				},
			},
		},
	}
	presenceProvider := &fakePresenceProvider{presence: agentdomain.NewAgentPresenceIndex()}

	svc := NewService(installed, sources, presenceProvider, viewstate.NewManager(t.TempDir()))
	input := StarSkillsInput{RepoScanMaxDepth: 2}
	repoURL := "https://github.com/acme/skills"

	first, err := svc.ListRepoStarSkills(context.Background(), repoURL, input)
	require.NoError(t, err)
	require.Len(t, first, 1)
	assert.True(t, first[0].Installed)

	second, err := svc.ListRepoStarSkills(context.Background(), repoURL, input)
	require.NoError(t, err)
	require.Len(t, second, 1)

	assert.Equal(t, 2, sources.repoCalls[repoURL])
	assert.Equal(t, 2, installed.calls)
	assert.Equal(t, 2, presenceProvider.calls)
}

func TestServicePropagatesProviderErrors(t *testing.T) {
	want := errors.New("boom")
	svc := NewService(
		&fakeInstalledProvider{err: want},
		&fakeSourceProvider{},
		&fakePresenceProvider{presence: agentdomain.NewAgentPresenceIndex()},
		nil,
	)

	_, err := svc.ListInstalledSkills(context.Background(), InstalledSkillsInput{})
	require.Error(t, err)
	assert.ErrorIs(t, err, want)
}
