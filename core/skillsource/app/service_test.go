package app

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	sourcedomain "github.com/shinerio/skillflow/core/skillsource/domain"
)

type fakeStorage struct {
	repos []sourcedomain.StarRepo
}

func (f *fakeStorage) Load() ([]sourcedomain.StarRepo, error) {
	return append([]sourcedomain.StarRepo(nil), f.repos...), nil
}

func (f *fakeStorage) Save(repos []sourcedomain.StarRepo) error {
	f.repos = append([]sourcedomain.StarRepo(nil), repos...)
	return nil
}

type fakeScanner struct {
	byDir map[string][]sourcedomain.SourceSkillCandidate
}

func (f *fakeScanner) ScanSkillsWithMaxDepth(repoDir, repoURL, repoName, source string, maxDepth int) ([]sourcedomain.SourceSkillCandidate, error) {
	return append([]sourcedomain.SourceSkillCandidate(nil), f.byDir[repoDir]...), nil
}

type fakeGitClient struct {
	nameByURL       map[string]string
	sourceByURL     map[string]string
	cacheByURL      map[string]string
	cloneCalls      []string
	cloneErrByURL   map[string]error
	cloneCredsCalls []string
	authError       error
	sshAuthError    error
}

func (f *fakeGitClient) CheckGitInstalled() error { return nil }
func (f *fakeGitClient) ParseRepoName(repoURL string) (string, error) {
	return f.nameByURL[repoURL], nil
}
func (f *fakeGitClient) RepoSource(repoURL string) (string, error) {
	return f.sourceByURL[repoURL], nil
}
func (f *fakeGitClient) CacheDir(dataDir, repoURL string) (string, error) {
	return f.cacheByURL[repoURL], nil
}
func (f *fakeGitClient) SameRepo(repoA, repoB string) bool {
	return f.sourceByURL[repoA] == f.sourceByURL[repoB]
}
func (f *fakeGitClient) CloneOrUpdate(ctx context.Context, repoURL, dir, proxyURL string) error {
	f.cloneCalls = append(f.cloneCalls, repoURL)
	return f.cloneErrByURL[repoURL]
}
func (f *fakeGitClient) CloneOrUpdateWithCreds(ctx context.Context, repoURL, dir, proxyURL, username, password string) error {
	f.cloneCredsCalls = append(f.cloneCredsCalls, repoURL)
	return f.cloneErrByURL[repoURL]
}
func (f *fakeGitClient) IsAuthError(err error) bool    { return err == f.authError }
func (f *fakeGitClient) IsSSHAuthError(err error) bool { return err == f.sshAuthError }

func TestTrackStarRepoAddsNewTrackedRepo(t *testing.T) {
	storage := &fakeStorage{}
	git := &fakeGitClient{
		nameByURL:   map[string]string{"https://github.com/octo/demo": "octo/demo"},
		sourceByURL: map[string]string{"https://github.com/octo/demo": "github.com/octo/demo"},
		cacheByURL:  map[string]string{"https://github.com/octo/demo": filepath.Join("/data", "cache", "github.com", "octo", "demo")},
	}
	service := NewService(storage, &fakeScanner{}, git)

	repo, err := service.TrackStarRepo(context.Background(), "/data", "https://github.com/octo/demo", "")
	if err != nil {
		t.Fatal(err)
	}

	if repo.URL != "https://github.com/octo/demo" {
		t.Fatalf("unexpected repo: %+v", repo)
	}
	if repo.Name != "octo/demo" || repo.Source != "github.com/octo/demo" {
		t.Fatalf("unexpected repo metadata: %+v", repo)
	}
	if repo.LocalDir != filepath.Join("/data", "cache", "github.com", "octo", "demo") {
		t.Fatalf("unexpected local dir: %+v", repo)
	}
	if repo.LastSync.IsZero() {
		t.Fatalf("expected last sync after successful clone, got %+v", repo)
	}
	if len(storage.repos) != 1 {
		t.Fatalf("expected 1 repo persisted, got %+v", storage.repos)
	}
	if len(git.cloneCalls) != 1 || git.cloneCalls[0] != "https://github.com/octo/demo" {
		t.Fatalf("unexpected clone calls: %+v", git.cloneCalls)
	}
}

func TestTrackStarRepoReturnsExistingTrackedRepoWithoutCloning(t *testing.T) {
	storage := &fakeStorage{repos: []sourcedomain.StarRepo{{
		URL:      "https://github.com/octo/demo.git",
		Name:     "octo/demo",
		Source:   "github.com/octo/demo",
		LocalDir: "/data/cache/github.com/octo/demo",
	}}}
	git := &fakeGitClient{
		sourceByURL: map[string]string{
			"https://github.com/octo/demo.git": "github.com/octo/demo",
			"https://github.com/octo/demo":     "github.com/octo/demo",
		},
	}
	service := NewService(storage, &fakeScanner{}, git)

	repo, err := service.TrackStarRepo(context.Background(), "/data", "https://github.com/octo/demo", "")
	if err != nil {
		t.Fatal(err)
	}
	if repo.URL != "https://github.com/octo/demo.git" {
		t.Fatalf("expected existing repo returned, got %+v", repo)
	}
	if len(git.cloneCalls) != 0 {
		t.Fatalf("expected no clone, got %+v", git.cloneCalls)
	}
}

func TestTrackStarRepoRecordsSyncErrorWhenCloneFails(t *testing.T) {
	storage := &fakeStorage{}
	git := &fakeGitClient{
		nameByURL:     map[string]string{"https://github.com/octo/demo": "octo/demo"},
		sourceByURL:   map[string]string{"https://github.com/octo/demo": "github.com/octo/demo"},
		cacheByURL:    map[string]string{"https://github.com/octo/demo": filepath.Join("/data", "cache", "github.com", "octo", "demo")},
		cloneErrByURL: map[string]error{"https://github.com/octo/demo": errors.New("network timeout")},
	}
	service := NewService(storage, &fakeScanner{}, git)

	repo, err := service.TrackStarRepo(context.Background(), "/data", "https://github.com/octo/demo", "")
	if err != nil {
		t.Fatal(err)
	}
	if repo.SyncError != "network timeout" {
		t.Fatalf("expected sync error to be recorded, got %+v", repo)
	}
	if !repo.LastSync.IsZero() {
		t.Fatalf("expected zero last sync on failure, got %+v", repo)
	}
}

func TestTrackStarRepoReturnsAuthErrorWithoutPersistingFailedRepo(t *testing.T) {
	authErr := errors.New("authentication failed")
	storage := &fakeStorage{}
	git := &fakeGitClient{
		nameByURL:     map[string]string{"https://github.com/octo/demo": "octo/demo"},
		sourceByURL:   map[string]string{"https://github.com/octo/demo": "github.com/octo/demo"},
		cacheByURL:    map[string]string{"https://github.com/octo/demo": filepath.Join("/data", "cache", "github.com", "octo", "demo")},
		cloneErrByURL: map[string]error{"https://github.com/octo/demo": authErr},
		authError:     authErr,
	}
	service := NewService(storage, &fakeScanner{}, git)

	repo, err := service.TrackStarRepo(context.Background(), "/data", "https://github.com/octo/demo", "")
	if err == nil {
		t.Fatal("expected auth error, got nil")
	}
	if repo != nil {
		t.Fatalf("expected nil repo on auth error, got %+v", repo)
	}
	if len(storage.repos) != 0 {
		t.Fatalf("expected failed auth repo not persisted, got %+v", storage.repos)
	}
}

func TestTrackStarRepoWithCredentialsReplacesFailedEntry(t *testing.T) {
	storage := &fakeStorage{repos: []sourcedomain.StarRepo{{
		URL:       "https://github.com/octo/demo",
		Name:      "octo/demo",
		Source:    "github.com/octo/demo",
		LocalDir:  "/old",
		SyncError: "auth failed",
	}}}
	git := &fakeGitClient{
		nameByURL:   map[string]string{"https://github.com/octo/demo": "octo/demo"},
		sourceByURL: map[string]string{"https://github.com/octo/demo": "github.com/octo/demo"},
		cacheByURL:  map[string]string{"https://github.com/octo/demo": "/new"},
	}
	service := NewService(storage, &fakeScanner{}, git)

	repo, err := service.TrackStarRepoWithCredentials(context.Background(), "/data", "https://github.com/octo/demo", "", "user", "pass")
	if err != nil {
		t.Fatal(err)
	}
	if repo.LocalDir != "/new" {
		t.Fatalf("expected replaced repo, got %+v", repo)
	}
	if len(storage.repos) != 1 {
		t.Fatalf("expected one repo after replace, got %+v", storage.repos)
	}
	if len(git.cloneCredsCalls) != 1 {
		t.Fatalf("expected credentialed clone call, got %+v", git.cloneCredsCalls)
	}
}

func TestUntrackStarRepoRemovesMatchingRepo(t *testing.T) {
	storage := &fakeStorage{repos: []sourcedomain.StarRepo{
		{URL: "https://github.com/octo/demo", Source: "github.com/octo/demo"},
		{URL: "https://github.com/octo/other", Source: "github.com/octo/other"},
	}}
	git := &fakeGitClient{
		sourceByURL: map[string]string{
			"https://github.com/octo/demo":  "github.com/octo/demo",
			"https://github.com/octo/other": "github.com/octo/other",
		},
	}
	service := NewService(storage, &fakeScanner{}, git)

	if err := service.UntrackStarRepo("https://github.com/octo/demo"); err != nil {
		t.Fatal(err)
	}
	if len(storage.repos) != 1 || storage.repos[0].URL != "https://github.com/octo/other" {
		t.Fatalf("unexpected persisted repos: %+v", storage.repos)
	}
}

func TestListAllSourceCandidatesAggregatesTrackedRepos(t *testing.T) {
	storage := &fakeStorage{repos: []sourcedomain.StarRepo{
		{URL: "https://github.com/octo/demo", Name: "octo/demo", Source: "github.com/octo/demo", LocalDir: "/demo"},
		{URL: "https://github.com/octo/other", Name: "octo/other", Source: "github.com/octo/other", LocalDir: "/other"},
	}}
	scanner := &fakeScanner{byDir: map[string][]sourcedomain.SourceSkillCandidate{
		"/demo":  {{Name: "demo-skill", Path: "/demo/skills/demo-skill", RepoURL: "https://github.com/octo/demo", RepoName: "octo/demo", Source: "github.com/octo/demo", SubPath: "skills/demo-skill"}},
		"/other": {{Name: "other-skill", Path: "/other/skills/other-skill", RepoURL: "https://github.com/octo/other", RepoName: "octo/other", Source: "github.com/octo/other", SubPath: "skills/other-skill"}},
	}}
	service := NewService(storage, scanner, &fakeGitClient{})

	candidates, err := service.ListAllSourceCandidates(5)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %+v", candidates)
	}
}

func TestRefreshStarRepoUpdatesSyncState(t *testing.T) {
	storage := &fakeStorage{repos: []sourcedomain.StarRepo{{
		URL:      "https://github.com/octo/demo",
		Name:     "octo/demo",
		Source:   "github.com/octo/demo",
		LocalDir: "/demo",
	}}}
	service := NewService(storage, &fakeScanner{}, &fakeGitClient{})

	repo, err := service.RefreshStarRepo(context.Background(), "https://github.com/octo/demo", "")
	if err != nil {
		t.Fatal(err)
	}
	if repo.SyncError != "" || repo.LastSync.IsZero() {
		t.Fatalf("expected successful refresh, got %+v", repo)
	}
}

func TestRefreshAllStarReposRecordsFailuresAndSuccesses(t *testing.T) {
	storage := &fakeStorage{repos: []sourcedomain.StarRepo{
		{URL: "https://github.com/octo/demo", Name: "octo/demo", Source: "github.com/octo/demo", LocalDir: "/demo"},
		{URL: "https://github.com/octo/bad", Name: "octo/bad", Source: "github.com/octo/bad", LocalDir: "/bad"},
	}}
	git := &fakeGitClient{
		cloneErrByURL: map[string]error{
			"https://github.com/octo/bad": errors.New("forced sync failure"),
		},
	}
	service := NewService(storage, &fakeScanner{}, git)

	results, err := service.RefreshAllStarRepos(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %+v", results)
	}
	if storage.repos[0].LastSync.IsZero() {
		t.Fatalf("expected successful repo to have last sync, got %+v", storage.repos[0])
	}
	if storage.repos[1].SyncError != "forced sync failure" {
		t.Fatalf("expected failed repo to persist sync error, got %+v", storage.repos[1])
	}
	if storage.repos[1].LastSync.After(time.Now().Add(time.Second)) {
		t.Fatalf("unexpected last sync on failed repo: %+v", storage.repos[1])
	}
}
