package app

import (
	"context"
	"strings"
	"time"

	platformgit "github.com/shinerio/skillflow/core/platform/git"
	sourcedomain "github.com/shinerio/skillflow/core/skillsource/domain"
)

type StarRepoStore interface {
	Load() ([]sourcedomain.StarRepo, error)
	Save([]sourcedomain.StarRepo) error
}

type RepoScanner interface {
	ScanSkillsWithMaxDepth(repoDir, repoURL, repoName, source string, maxDepth int) ([]sourcedomain.SourceSkillCandidate, error)
}

type GitClient interface {
	CheckGitInstalled() error
	ParseRepoName(repoURL string) (string, error)
	RepoSource(repoURL string) (string, error)
	CacheDir(dataDir, repoURL string) (string, error)
	SameRepo(repoA, repoB string) bool
	CloneOrUpdate(ctx context.Context, repoURL, dir, proxyURL string) error
	CloneOrUpdateWithCreds(ctx context.Context, repoURL, dir, proxyURL, username, password string) error
	IsAuthError(err error) bool
	IsSSHAuthError(err error) bool
}

type Service struct {
	store   StarRepoStore
	scanner RepoScanner
	git     GitClient
	now     func() time.Time
}

type RefreshResult struct {
	Repo sourcedomain.StarRepo
}

func NewService(store StarRepoStore, scanner RepoScanner, git GitClient) *Service {
	return &Service{
		store:   store,
		scanner: scanner,
		git:     git,
		now:     time.Now,
	}
}

type platformGitClient struct{}

func NewPlatformGitClient() GitClient {
	return platformGitClient{}
}

func (platformGitClient) CheckGitInstalled() error {
	return platformgit.CheckGitInstalled()
}

func (platformGitClient) ParseRepoName(repoURL string) (string, error) {
	return platformgit.ParseRepoName(repoURL)
}

func (platformGitClient) RepoSource(repoURL string) (string, error) {
	return platformgit.RepoSource(repoURL)
}

func (platformGitClient) CacheDir(dataDir, repoURL string) (string, error) {
	return platformgit.CacheDir(dataDir, repoURL)
}

func (platformGitClient) SameRepo(repoA, repoB string) bool {
	return platformgit.SameRepo(repoA, repoB)
}

func (platformGitClient) CloneOrUpdate(ctx context.Context, repoURL, dir, proxyURL string) error {
	return platformgit.CloneOrUpdate(ctx, repoURL, dir, proxyURL)
}

func (platformGitClient) CloneOrUpdateWithCreds(ctx context.Context, repoURL, dir, proxyURL, username, password string) error {
	return platformgit.CloneOrUpdateWithCreds(ctx, repoURL, dir, proxyURL, username, password)
}

func (platformGitClient) IsAuthError(err error) bool {
	return platformgit.IsAuthError(err)
}

func (platformGitClient) IsSSHAuthError(err error) bool {
	return platformgit.IsSSHAuthError(err)
}

func (s *Service) TrackStarRepo(ctx context.Context, dataDir, repoURL, proxyURL string) (*sourcedomain.StarRepo, error) {
	if err := s.git.CheckGitInstalled(); err != nil {
		return nil, err
	}
	repos, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	for i := range repos {
		if s.git.SameRepo(repos[i].URL, repoURL) {
			if repos[i].Source == "" {
				if source, err := s.git.RepoSource(repos[i].URL); err == nil {
					repos[i].Source = source
					_ = s.store.Save(repos)
				}
			}
			return &repos[i], nil
		}
	}
	repo, err := s.newRepo(dataDir, repoURL)
	if err != nil {
		return nil, err
	}
	if cloneErr := s.git.CloneOrUpdate(ctx, repoURL, repo.LocalDir, proxyURL); cloneErr != nil {
		if s.git.IsSSHAuthError(cloneErr) || s.git.IsAuthError(cloneErr) {
			return nil, cloneErr
		}
		repo.SyncError = cloneErr.Error()
	} else {
		repo.LastSync = s.now()
	}
	repos = append(repos, repo)
	if err := s.store.Save(repos); err != nil {
		return nil, err
	}
	return &repos[len(repos)-1], nil
}

func (s *Service) TrackStarRepoWithCredentials(ctx context.Context, dataDir, repoURL, proxyURL, username, password string) (*sourcedomain.StarRepo, error) {
	if err := s.git.CheckGitInstalled(); err != nil {
		return nil, err
	}
	repos, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	filtered := repos[:0]
	for _, repo := range repos {
		if !s.git.SameRepo(repo.URL, repoURL) {
			filtered = append(filtered, repo)
		}
	}
	repo, err := s.newRepo(dataDir, repoURL)
	if err != nil {
		return nil, err
	}
	if err := s.git.CloneOrUpdateWithCreds(ctx, repoURL, repo.LocalDir, proxyURL, username, password); err != nil {
		return nil, err
	}
	repo.LastSync = s.now()
	filtered = append(filtered, repo)
	if err := s.store.Save(filtered); err != nil {
		return nil, err
	}
	return &filtered[len(filtered)-1], nil
}

func (s *Service) UntrackStarRepo(repoURL string) error {
	repos, err := s.store.Load()
	if err != nil {
		return err
	}
	filtered := make([]sourcedomain.StarRepo, 0, len(repos))
	for _, repo := range repos {
		if !s.git.SameRepo(repo.URL, repoURL) {
			filtered = append(filtered, repo)
		}
	}
	return s.store.Save(filtered)
}

func (s *Service) ListStarRepos() ([]sourcedomain.StarRepo, error) {
	repos, err := s.store.Load()
	if repos == nil {
		return []sourcedomain.StarRepo{}, err
	}
	changed := false
	for i := range repos {
		if repos[i].Source != "" {
			continue
		}
		if source, parseErr := s.git.RepoSource(repos[i].URL); parseErr == nil {
			repos[i].Source = source
			changed = true
		}
	}
	if changed {
		_ = s.store.Save(repos)
	}
	return repos, err
}

func (s *Service) ListAllSourceCandidates(maxDepth int) ([]sourcedomain.SourceSkillCandidate, error) {
	repos, err := s.ListStarRepos()
	if err != nil {
		return nil, err
	}
	var all []sourcedomain.SourceSkillCandidate
	for _, repo := range repos {
		candidates, err := s.scanner.ScanSkillsWithMaxDepth(repo.LocalDir, repo.URL, repo.Name, repo.Source, maxDepth)
		if err != nil {
			return nil, err
		}
		all = append(all, candidates...)
	}
	if all == nil {
		return []sourcedomain.SourceSkillCandidate{}, nil
	}
	return all, nil
}

func (s *Service) ListSourceCandidatesByRepo(repoURL string, maxDepth int) ([]sourcedomain.SourceSkillCandidate, error) {
	repos, err := s.ListStarRepos()
	if err != nil {
		return nil, err
	}
	for _, repo := range repos {
		if !s.git.SameRepo(repo.URL, repoURL) {
			continue
		}
		candidates, err := s.scanner.ScanSkillsWithMaxDepth(repo.LocalDir, repo.URL, repo.Name, repo.Source, maxDepth)
		if err != nil {
			return nil, err
		}
		if candidates == nil {
			return []sourcedomain.SourceSkillCandidate{}, nil
		}
		return candidates, nil
	}
	return []sourcedomain.SourceSkillCandidate{}, nil
}

func (s *Service) RefreshStarRepo(ctx context.Context, repoURL, proxyURL string) (*sourcedomain.StarRepo, error) {
	repos, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	for i, repo := range repos {
		if !s.git.SameRepo(repo.URL, repoURL) {
			continue
		}
		if syncErr := s.git.CloneOrUpdate(ctx, repo.URL, repo.LocalDir, proxyURL); syncErr != nil {
			repos[i].SyncError = syncErr.Error()
		} else {
			repos[i].SyncError = ""
			repos[i].LastSync = s.now()
		}
		if err := s.store.Save(repos); err != nil {
			return nil, err
		}
		return &repos[i], nil
	}
	return nil, nil
}

func (s *Service) RefreshAllStarRepos(ctx context.Context, proxyURL string) ([]RefreshResult, error) {
	repos, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	results := make([]RefreshResult, 0, len(repos))
	for i := range repos {
		if syncErr := s.git.CloneOrUpdate(ctx, repos[i].URL, repos[i].LocalDir, proxyURL); syncErr != nil {
			repos[i].SyncError = syncErr.Error()
		} else {
			repos[i].SyncError = ""
			repos[i].LastSync = s.now()
		}
		results = append(results, RefreshResult{Repo: repos[i]})
	}
	if err := s.store.Save(repos); err != nil {
		return nil, err
	}
	return results, nil
}

func (s *Service) newRepo(dataDir, repoURL string) (sourcedomain.StarRepo, error) {
	name, err := s.git.ParseRepoName(repoURL)
	if err != nil {
		return sourcedomain.StarRepo{}, err
	}
	localDir, err := s.git.CacheDir(dataDir, repoURL)
	if err != nil {
		return sourcedomain.StarRepo{}, err
	}
	source, err := s.git.RepoSource(repoURL)
	if err != nil {
		return sourcedomain.StarRepo{}, err
	}
	return sourcedomain.StarRepo{
		URL:      strings.TrimSpace(repoURL),
		Name:     name,
		Source:   source,
		LocalDir: localDir,
	}, nil
}
