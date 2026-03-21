package orchestration

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
	skillquery "github.com/shinerio/skillflow/core/skillcatalog/app/query"
	skilldomain "github.com/shinerio/skillflow/core/skillcatalog/domain"
	sourcedomain "github.com/shinerio/skillflow/core/skillsource/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type importCall struct {
	srcDir   string
	category string
	source   skilldomain.SourceType
}

type overwriteCall struct {
	id     string
	srcDir string
}

type fakeSkillCatalog struct {
	importFn       func(srcDir, category string, source skilldomain.SourceType, sourceURL, sourceSubPath string) (*skilldomain.InstalledSkill, error)
	getFn          func(id string) (*skilldomain.InstalledSkill, error)
	listFn         func() ([]*skilldomain.InstalledSkill, error)
	deleteFn       func(id string) error
	overwriteFn    func(id, srcDir string) error
	updateMetaFn   func(sk *skilldomain.InstalledSkill) error
	importCalls    []importCall
	deleteCalls    []string
	overwriteCalls []overwriteCall
	updateCalls    []*skilldomain.InstalledSkill
}

func (f *fakeSkillCatalog) Import(srcDir, category string, source skilldomain.SourceType, sourceURL, sourceSubPath string) (*skilldomain.InstalledSkill, error) {
	f.importCalls = append(f.importCalls, importCall{srcDir: srcDir, category: category, source: source})
	if f.importFn != nil {
		return f.importFn(srcDir, category, source, sourceURL, sourceSubPath)
	}
	return nil, nil
}

func (f *fakeSkillCatalog) Get(id string) (*skilldomain.InstalledSkill, error) {
	if f.getFn != nil {
		return f.getFn(id)
	}
	return nil, nil
}

func (f *fakeSkillCatalog) ListAll() ([]*skilldomain.InstalledSkill, error) {
	if f.listFn != nil {
		return f.listFn()
	}
	return nil, nil
}

func (f *fakeSkillCatalog) Delete(id string) error {
	f.deleteCalls = append(f.deleteCalls, id)
	if f.deleteFn != nil {
		return f.deleteFn(id)
	}
	return nil
}

func (f *fakeSkillCatalog) OverwriteFromDir(id, srcDir string) error {
	f.overwriteCalls = append(f.overwriteCalls, overwriteCall{id: id, srcDir: srcDir})
	if f.overwriteFn != nil {
		return f.overwriteFn(id, srcDir)
	}
	return nil
}

func (f *fakeSkillCatalog) UpdateMeta(sk *skilldomain.InstalledSkill) error {
	copied := *sk
	f.updateCalls = append(f.updateCalls, &copied)
	if f.updateMetaFn != nil {
		return f.updateMetaFn(sk)
	}
	return nil
}

type pushCall struct {
	profiles []agentdomain.AgentProfile
	agents   []string
	skills   []*skilldomain.InstalledSkill
	force    bool
}

type fakeAgentIntegration struct {
	pushFn          func(ctx context.Context, profiles []agentdomain.AgentProfile, agentNames []string, skills []*skilldomain.InstalledSkill, force bool) ([]agentdomain.PushConflict, error)
	buildPresenceFn func(ctx context.Context, profiles []agentdomain.AgentProfile, idx *skillquery.InstalledIndex, maxDepth int) (*agentdomain.AgentPresenceIndex, error)
	scanFn          func(ctx context.Context, profile agentdomain.AgentProfile, idx *skillquery.InstalledIndex, presence *agentdomain.AgentPresenceIndex, maxDepth int) ([]agentdomain.AgentSkillCandidate, error)
	refreshFn       func(ctx context.Context, profiles []agentdomain.AgentProfile, skill *skilldomain.InstalledSkill) error
	pushCalls       []pushCall
	refreshCalls    int
}

func (f *fakeAgentIntegration) PushSkills(ctx context.Context, profiles []agentdomain.AgentProfile, agentNames []string, skills []*skilldomain.InstalledSkill, force bool) ([]agentdomain.PushConflict, error) {
	copiedProfiles := append([]agentdomain.AgentProfile(nil), profiles...)
	copiedAgents := append([]string(nil), agentNames...)
	copiedSkills := append([]*skilldomain.InstalledSkill(nil), skills...)
	f.pushCalls = append(f.pushCalls, pushCall{
		profiles: copiedProfiles,
		agents:   copiedAgents,
		skills:   copiedSkills,
		force:    force,
	})
	if f.pushFn != nil {
		return f.pushFn(ctx, profiles, agentNames, skills, force)
	}
	return nil, nil
}

func (f *fakeAgentIntegration) BuildPresenceIndex(ctx context.Context, profiles []agentdomain.AgentProfile, idx *skillquery.InstalledIndex, maxDepth int) (*agentdomain.AgentPresenceIndex, error) {
	if f.buildPresenceFn != nil {
		return f.buildPresenceFn(ctx, profiles, idx, maxDepth)
	}
	return agentdomain.NewAgentPresenceIndex(), nil
}

func (f *fakeAgentIntegration) ScanAgentSkills(ctx context.Context, profile agentdomain.AgentProfile, idx *skillquery.InstalledIndex, presence *agentdomain.AgentPresenceIndex, maxDepth int) ([]agentdomain.AgentSkillCandidate, error) {
	if f.scanFn != nil {
		return f.scanFn(ctx, profile, idx, presence, maxDepth)
	}
	return nil, nil
}

func (f *fakeAgentIntegration) RefreshPushedCopies(ctx context.Context, profiles []agentdomain.AgentProfile, skill *skilldomain.InstalledSkill) error {
	f.refreshCalls++
	if f.refreshFn != nil {
		return f.refreshFn(ctx, profiles, skill)
	}
	return nil
}

type fakeSkillSourceResolver struct {
	sourceDir string
	latestSHA string
	err       error
}

func (f fakeSkillSourceResolver) ResolveCachedSource(context.Context, *skilldomain.InstalledSkill) (string, string, error) {
	return f.sourceDir, f.latestSHA, f.err
}

type fakeBackupScheduler struct {
	calls []string
	err   error
}

func (f *fakeBackupScheduler) ScheduleAutoBackup(_ context.Context, source string) error {
	f.calls = append(f.calls, source)
	return f.err
}

type fakeStarRepoStore struct {
	repos      []sourcedomain.StarRepo
	saveCalls  int
	savedRepos []sourcedomain.StarRepo
}

func (f *fakeStarRepoStore) Load() ([]sourcedomain.StarRepo, error) {
	return append([]sourcedomain.StarRepo(nil), f.repos...), nil
}

func (f *fakeStarRepoStore) Save(repos []sourcedomain.StarRepo) error {
	f.saveCalls++
	f.savedRepos = append([]sourcedomain.StarRepo(nil), repos...)
	f.repos = append([]sourcedomain.StarRepo(nil), repos...)
	return nil
}

type cloneCall struct {
	repoURL  string
	dir      string
	proxyURL string
}

type fakeRepoCloner struct {
	errByURL map[string]error
	calls    []cloneCall
}

func (f *fakeRepoCloner) CloneOrUpdate(_ context.Context, repoURL, dir, proxyURL string) error {
	f.calls = append(f.calls, cloneCall{repoURL: repoURL, dir: dir, proxyURL: proxyURL})
	if f.errByURL != nil {
		return f.errByURL[repoURL]
	}
	return nil
}

func TestImportLocalSkill_AutoPushAndBackup(t *testing.T) {
	catalog := &fakeSkillCatalog{
		importFn: func(srcDir, category string, source skilldomain.SourceType, sourceURL, sourceSubPath string) (*skilldomain.InstalledSkill, error) {
			return &skilldomain.InstalledSkill{
				ID:       "skill-1",
				Name:     "demo-skill",
				Path:     srcDir,
				Category: category,
				Source:   source,
			}, nil
		},
	}
	agents := &fakeAgentIntegration{}
	backup := &fakeBackupScheduler{}
	svc := NewService(Dependencies{
		SkillCatalog:     catalog,
		AgentIntegration: agents,
		AutoBackup:       backup,
	})

	profiles := []agentdomain.AgentProfile{
		{Name: "codex", Enabled: true},
		{Name: "claude", Enabled: false},
	}
	result, err := svc.ImportLocalSkill(context.Background(), ImportLocalCommand{
		SourceDir:          "/tmp/demo-skill",
		Category:           "",
		AgentProfiles:      profiles,
		AutoPushAgentNames: []string{"codex", "claude"},
		TriggerAutoBackup:  true,
	})
	require.NoError(t, err)
	require.NotNil(t, result.Skill)
	require.Len(t, catalog.importCalls, 1)
	assert.Equal(t, "Default", catalog.importCalls[0].category)
	assert.Equal(t, skilldomain.SourceManual, catalog.importCalls[0].source)
	require.Len(t, agents.pushCalls, 1)
	assert.Equal(t, []string{"codex"}, agents.pushCalls[0].agents)
	assert.False(t, agents.pushCalls[0].force)
	assert.Equal(t, []string{"local.import"}, backup.calls)
}

func TestImportRepoSourceSkills_SetsSourceMetadataAndRunsSideEffects(t *testing.T) {
	catalog := &fakeSkillCatalog{
		listFn: func() ([]*skilldomain.InstalledSkill, error) {
			return []*skilldomain.InstalledSkill{}, nil
		},
		importFn: func(srcDir, category string, source skilldomain.SourceType, sourceURL, sourceSubPath string) (*skilldomain.InstalledSkill, error) {
			return &skilldomain.InstalledSkill{
				ID:            "skill-1",
				Name:          "demo-skill",
				Path:          srcDir,
				Category:      category,
				Source:        source,
				SourceURL:     sourceURL,
				SourceSubPath: sourceSubPath,
			}, nil
		},
	}
	agents := &fakeAgentIntegration{}
	backup := &fakeBackupScheduler{}
	svc := NewService(Dependencies{
		SkillCatalog:     catalog,
		AgentIntegration: agents,
		AutoBackup:       backup,
		ResolveRepoSubPathSHA: func(_ context.Context, repoDir, subPath string) (string, error) {
			assert.Equal(t, "/tmp/cache/demo", repoDir)
			assert.Equal(t, "skills/demo-skill", subPath)
			return "sha-new", nil
		},
	})

	result, err := svc.ImportRepoSourceSkills(context.Background(), ImportRepoSourceSkillsCommand{
		SkillPaths:         []string{"/tmp/cache/demo/skills/demo-skill"},
		RepoRootDir:        "/tmp/cache/demo",
		CanonicalRepoURL:   "https://github.com/acme/demo",
		Category:           "",
		AgentProfiles:      []agentdomain.AgentProfile{{Name: "codex", Enabled: true}},
		AutoPushAgentNames: []string{"codex"},
		TriggerAutoBackup:  true,
	})
	require.NoError(t, err)
	require.Len(t, result.Imported, 1)
	assert.Equal(t, "sha-new", result.Imported[0].SourceSHA)
	require.Len(t, catalog.importCalls, 1)
	assert.Equal(t, skilldomain.SourceGitHub, catalog.importCalls[0].source)
	assert.Equal(t, "Default", catalog.importCalls[0].category)
	require.Len(t, catalog.updateCalls, 1)
	assert.Equal(t, "sha-new", catalog.updateCalls[0].SourceSHA)
	require.Len(t, agents.pushCalls, 1)
	assert.Equal(t, []string{"codex"}, agents.pushCalls[0].agents)
	assert.Equal(t, []string{"starred.import"}, backup.calls)
}

func TestPullFromAgent_NonForceCollectsConflictsAndRunsSideEffects(t *testing.T) {
	errSkillExists := errors.New("skill exists")
	catalog := &fakeSkillCatalog{
		listFn: func() ([]*skilldomain.InstalledSkill, error) {
			return []*skilldomain.InstalledSkill{}, nil
		},
		importFn: func(srcDir, category string, source skilldomain.SourceType, sourceURL, sourceSubPath string) (*skilldomain.InstalledSkill, error) {
			switch srcDir {
			case "/agent/beta":
				return &skilldomain.InstalledSkill{ID: "beta-id", Name: "beta"}, nil
			case "/agent/gamma":
				return nil, errSkillExists
			default:
				return nil, nil
			}
		},
	}
	agents := &fakeAgentIntegration{
		scanFn: func(ctx context.Context, profile agentdomain.AgentProfile, idx *skillquery.InstalledIndex, presence *agentdomain.AgentPresenceIndex, maxDepth int) ([]agentdomain.AgentSkillCandidate, error) {
			return []agentdomain.AgentSkillCandidate{
				{Name: "alpha", Path: "/agent/alpha", Imported: true},
				{Name: "beta", Path: "/agent/beta", Imported: false},
				{Name: "gamma", Path: "/agent/gamma", Imported: false},
			}, nil
		},
	}
	backup := &fakeBackupScheduler{}
	svc := NewService(Dependencies{
		SkillCatalog:     catalog,
		AgentIntegration: agents,
		AutoBackup:       backup,
		IsSkillExistsError: func(err error) bool {
			return errors.Is(err, errSkillExists)
		},
	})

	result, err := svc.PullFromAgent(context.Background(), PullFromAgentCommand{
		AgentName:          "codex",
		SkillPaths:         []string{"/agent/alpha", "/agent/beta", "/agent/gamma"},
		Category:           "Tools",
		AgentProfiles:      []agentdomain.AgentProfile{{Name: "codex", Enabled: true}},
		AutoPushAgentNames: []string{"codex"},
		TriggerAutoBackup:  true,
	})
	require.NoError(t, err)
	assert.True(t, result.AgentFound)
	assert.Equal(t, []string{"/agent/alpha", "/agent/gamma"}, result.Conflicts)
	require.Len(t, result.Imported, 1)
	assert.Equal(t, "beta-id", result.Imported[0].ID)
	require.Len(t, agents.pushCalls, 1)
	assert.False(t, agents.pushCalls[0].force)
	require.Len(t, agents.pushCalls[0].skills, 1)
	assert.Equal(t, "beta-id", agents.pushCalls[0].skills[0].ID)
	assert.Equal(t, []string{"agent.pull"}, backup.calls)
}

func TestPullFromAgent_ForceDeletesExistingAndAutoPushesWithOverwrite(t *testing.T) {
	existingGitHub := &skilldomain.InstalledSkill{
		ID:            "installed-1",
		Name:          "alpha",
		Source:        skilldomain.SourceGitHub,
		SourceURL:     "https://github.com/acme/alpha",
		SourceSubPath: "skills/alpha",
	}
	existingManual := &skilldomain.InstalledSkill{
		ID:   "installed-2",
		Name: "beta",
	}
	logicalKey, err := skilldomain.LogicalKey(existingGitHub)
	require.NoError(t, err)

	catalog := &fakeSkillCatalog{
		listFn: func() ([]*skilldomain.InstalledSkill, error) {
			return []*skilldomain.InstalledSkill{existingGitHub, existingManual}, nil
		},
		importFn: func(srcDir, category string, source skilldomain.SourceType, sourceURL, sourceSubPath string) (*skilldomain.InstalledSkill, error) {
			return &skilldomain.InstalledSkill{ID: filepath.Base(srcDir), Name: filepath.Base(srcDir)}, nil
		},
	}
	agents := &fakeAgentIntegration{
		scanFn: func(ctx context.Context, profile agentdomain.AgentProfile, idx *skillquery.InstalledIndex, presence *agentdomain.AgentPresenceIndex, maxDepth int) ([]agentdomain.AgentSkillCandidate, error) {
			return []agentdomain.AgentSkillCandidate{
				{Name: "alpha", Path: "/agent/new-alpha", LogicalKey: logicalKey},
				{Name: "beta", Path: "/agent/new-beta"},
			}, nil
		},
	}
	svc := NewService(Dependencies{
		SkillCatalog:     catalog,
		AgentIntegration: agents,
	})

	result, err := svc.PullFromAgent(context.Background(), PullFromAgentCommand{
		AgentName:          "codex",
		SkillPaths:         []string{"/agent/new-alpha", "/agent/new-beta"},
		Category:           "",
		Force:              true,
		AgentProfiles:      []agentdomain.AgentProfile{{Name: "codex", Enabled: true}},
		AutoPushAgentNames: []string{"codex"},
	})
	require.NoError(t, err)
	assert.Empty(t, result.Conflicts)
	assert.ElementsMatch(t, []string{"installed-1", "installed-2"}, catalog.deleteCalls)
	require.Len(t, agents.pushCalls, 1)
	assert.True(t, agents.pushCalls[0].force)
	assert.Len(t, agents.pushCalls[0].skills, 2)
}

func TestUpdateInstalledSkill_RefreshFailureStillTriggersBackup(t *testing.T) {
	updateErr := errors.New("refresh pushed copies failed")
	catalog := &fakeSkillCatalog{
		getFn: func(id string) (*skilldomain.InstalledSkill, error) {
			return &skilldomain.InstalledSkill{ID: id, Name: "demo", SourceSHA: "old", LatestSHA: "new"}, nil
		},
	}
	agents := &fakeAgentIntegration{
		refreshFn: func(ctx context.Context, profiles []agentdomain.AgentProfile, skill *skilldomain.InstalledSkill) error {
			return updateErr
		},
	}
	backup := &fakeBackupScheduler{}
	svc := NewService(Dependencies{
		SkillCatalog:     catalog,
		AgentIntegration: agents,
		SkillSource: fakeSkillSourceResolver{
			sourceDir: "/cache/demo",
			latestSHA: "latest-sha",
		},
		AutoBackup: backup,
	})

	_, err := svc.UpdateInstalledSkill(context.Background(), UpdateInstalledSkillCommand{
		SkillID:            "skill-1",
		AgentProfiles:      []agentdomain.AgentProfile{{Name: "codex", Enabled: true}},
		AutoPushAgentNames: []string{"codex"},
		TriggerAutoBackup:  true,
	})
	require.ErrorIs(t, err, updateErr)
	require.Len(t, catalog.overwriteCalls, 1)
	assert.Equal(t, "skill-1", catalog.overwriteCalls[0].id)
	assert.Equal(t, "/cache/demo", catalog.overwriteCalls[0].srcDir)
	require.Len(t, catalog.updateCalls, 1)
	assert.Equal(t, "latest-sha", catalog.updateCalls[0].SourceSHA)
	assert.Equal(t, "", catalog.updateCalls[0].LatestSHA)
	assert.Equal(t, []string{"skill.update"}, backup.calls)
	assert.Len(t, agents.pushCalls, 0)
}

func TestCompensateRestore_AutoPushesRestoredSkillsAndClonesNewRepos(t *testing.T) {
	now := time.Date(2026, 3, 21, 9, 0, 0, 0, time.UTC)
	unchanged := &skilldomain.InstalledSkill{ID: "keep", SourceSHA: "sha-1", UpdatedAt: now}
	restored := &skilldomain.InstalledSkill{ID: "new", SourceSHA: "sha-2", UpdatedAt: now.Add(time.Minute)}

	repoOldDir := t.TempDir()
	repoNewOKDir := t.TempDir()
	repoNewFailDir := t.TempDir()
	repoAlreadyClonedDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repoAlreadyClonedDir, ".git"), 0755))

	store := &fakeStarRepoStore{
		repos: []sourcedomain.StarRepo{
			{URL: "https://github.com/acme/old", LocalDir: repoOldDir},
			{URL: "https://github.com/acme/new-ok", LocalDir: repoNewOKDir},
			{URL: "https://github.com/acme/new-fail", LocalDir: repoNewFailDir},
			{URL: "https://github.com/acme/new-cloned", LocalDir: repoAlreadyClonedDir},
		},
	}
	cloner := &fakeRepoCloner{
		errByURL: map[string]error{
			"https://github.com/acme/new-fail": errors.New("network error"),
		},
	}
	catalog := &fakeSkillCatalog{
		listFn: func() ([]*skilldomain.InstalledSkill, error) {
			return []*skilldomain.InstalledSkill{unchanged, restored}, nil
		},
	}
	agents := &fakeAgentIntegration{}

	before := RestoreState{
		InstalledSkills: map[string]RestoreSkillSnapshot{
			restoreSkillKey(unchanged): {SourceSHA: unchanged.SourceSHA, UpdatedAt: unchanged.UpdatedAt},
		},
		StarredRepoURLs: map[string]struct{}{
			restoreRepoKey("https://github.com/acme/old"): {},
		},
	}

	svc := NewService(Dependencies{
		SkillCatalog:     catalog,
		AgentIntegration: agents,
		StarRepoStore:    store,
		RepoCloner:       cloner,
		Now: func() time.Time {
			return now.Add(2 * time.Minute)
		},
	})
	result, err := svc.CompensateRestore(context.Background(), RestoreCompensationCommand{
		Before:             before,
		Source:             "backup.restore",
		AgentProfiles:      []agentdomain.AgentProfile{{Name: "codex", Enabled: true}},
		AutoPushAgentNames: []string{"codex"},
		ProxyURL:           "http://proxy.local",
	})
	require.NoError(t, err)
	require.Len(t, result.RestoredSkills, 1)
	assert.Equal(t, "new", result.RestoredSkills[0].ID)
	assert.Equal(t, 1, result.ClonedRepos)
	assert.Equal(t, 1, result.FailedRepos)
	require.Len(t, agents.pushCalls, 1)
	assert.True(t, agents.pushCalls[0].force)
	require.Len(t, agents.pushCalls[0].skills, 1)
	assert.Equal(t, "new", agents.pushCalls[0].skills[0].ID)
	assert.Equal(t, 1, store.saveCalls)
	require.Len(t, cloner.calls, 2)
	assert.Equal(t, "https://github.com/acme/new-ok", cloner.calls[0].repoURL)
	assert.Equal(t, "http://proxy.local", cloner.calls[0].proxyURL)
	assert.Equal(t, "", store.savedRepos[1].SyncError)
	assert.Equal(t, "network error", store.savedRepos[2].SyncError)
	assert.False(t, store.savedRepos[1].LastSync.IsZero())
}

func TestCaptureRestoreState_UsesLogicalAndCanonicalKeys(t *testing.T) {
	gitSkill := &skilldomain.InstalledSkill{
		ID:            "git-skill",
		Source:        skilldomain.SourceGitHub,
		SourceURL:     "git@github.com:acme/demo.git",
		SourceSubPath: "skills/demo",
		SourceSHA:     "abc",
		UpdatedAt:     time.Date(2026, 3, 21, 1, 2, 3, 0, time.UTC),
	}
	manualSkill := &skilldomain.InstalledSkill{
		ID:        "manual-skill",
		Source:    skilldomain.SourceManual,
		SourceSHA: "def",
		UpdatedAt: time.Date(2026, 3, 21, 2, 3, 4, 0, time.UTC),
	}
	catalog := &fakeSkillCatalog{
		listFn: func() ([]*skilldomain.InstalledSkill, error) {
			return []*skilldomain.InstalledSkill{gitSkill, manualSkill}, nil
		},
	}
	store := &fakeStarRepoStore{
		repos: []sourcedomain.StarRepo{
			{URL: "git@github.com:acme/demo.git"},
		},
	}
	svc := NewService(Dependencies{
		SkillCatalog:  catalog,
		StarRepoStore: store,
	})

	state, err := svc.CaptureRestoreState(context.Background())
	require.NoError(t, err)
	gitKey := restoreSkillKey(gitSkill)
	manualKey := restoreSkillKey(manualSkill)
	_, hasGitSkill := state.InstalledSkills[gitKey]
	_, hasManualSkill := state.InstalledSkills[manualKey]
	_, hasRepo := state.StarredRepoURLs["https://github.com/acme/demo"]
	assert.True(t, hasGitSkill)
	assert.True(t, hasManualSkill)
	assert.True(t, hasRepo)
}
