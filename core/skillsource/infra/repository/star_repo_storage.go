package repository

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	platformgit "github.com/shinerio/skillflow/core/platform/git"
	"github.com/shinerio/skillflow/core/platform/pathutil"
	sourcedomain "github.com/shinerio/skillflow/core/skillsource/domain"
)

type StarRepoStorage struct {
	path            string
	localPath       string
	dataDir         string
	builtinRepoURLs []string
	mu              sync.Mutex
}

func NewStarRepoStorage(path string) *StarRepoStorage {
	return NewStarRepoStorageWithBuiltins(path, nil)
}

func NewStarRepoStorageWithBuiltins(path string, builtinRepoURLs []string) *StarRepoStorage {
	cleanPath := filepath.Clean(path)
	builtins := append([]string(nil), builtinRepoURLs...)
	dataDir := filepath.Dir(cleanPath)
	return &StarRepoStorage{
		path:            cleanPath,
		localPath:       filepath.Join(dataDir, "star_repos_local.json"),
		dataDir:         dataDir,
		builtinRepoURLs: builtins,
	}
}

func (s *StarRepoStorage) Load() ([]sourcedomain.StarRepo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		if len(s.builtinRepoURLs) == 0 {
			return nil, nil
		}
		repos, buildErr := s.buildBuiltinReposLocked()
		if buildErr != nil {
			return nil, buildErr
		}
		if err := s.saveLocked(repos); err != nil {
			return nil, err
		}
		return repos, nil
	}
	if err != nil {
		return nil, err
	}
	var repos []sourcedomain.StarRepo
	if err := json.Unmarshal(data, &repos); err != nil {
		return repos, err
	}
	localState, localErr := s.loadLocalStateLocked()
	if localErr != nil {
		return nil, localErr
	}
	changed := false
	for i := range repos {
		if s.resolveLocalDir(&repos[i]) {
			changed = true
		}
		if !repos[i].LastSync.IsZero() || strings.TrimSpace(repos[i].SyncError) != "" {
			changed = true
		}
		if local, ok := localState[s.localStateKey(repos[i])]; ok {
			repos[i].LastSync = local.LastSync
			repos[i].SyncError = local.SyncError
		}
	}
	if changed {
		if err := s.saveLocked(repos); err != nil {
			return nil, err
		}
	}
	return repos, nil
}

func (s *StarRepoStorage) Save(repos []sourcedomain.StarRepo) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked(repos)
}

func (s *StarRepoStorage) saveLocked(repos []sourcedomain.StarRepo) error {
	if repos == nil {
		repos = []sourcedomain.StarRepo{}
	}
	snapshot := make([]syncedStarRepo, len(repos))
	for i := range repos {
		snapshot[i] = s.serializedRepo(repos[i])
	}
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".star_repos_*.json")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		tmp.Close()
		os.Remove(tmpName)
	}()
	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		return err
	}
	return s.saveLocalStateLocked(repos)
}

type syncedStarRepo struct {
	URL      string `json:"url"`
	Name     string `json:"name"`
	Source   string `json:"source"`
	LocalDir string `json:"localDir"`
}

func (s *StarRepoStorage) serializedRepo(repo sourcedomain.StarRepo) syncedStarRepo {
	return syncedStarRepo{
		URL:      repo.URL,
		Name:     repo.Name,
		Source:   repo.Source,
		LocalDir: pathutil.StorePath(s.dataDir, repo.LocalDir, s.derivedLocalDir(repo.URL)),
	}
}

type localRepoState struct {
	LastSync  time.Time `json:"lastSync,omitempty"`
	SyncError string    `json:"syncError,omitempty"`
}

type starReposLocalSnapshot struct {
	Repos map[string]localRepoState `json:"repos"`
}

func (s *StarRepoStorage) localStateKey(repo sourcedomain.StarRepo) string {
	if source := strings.TrimSpace(repo.Source); source != "" {
		return strings.ToLower(source)
	}
	if source, err := platformgit.RepoSource(repo.URL); err == nil && strings.TrimSpace(source) != "" {
		return strings.ToLower(strings.TrimSpace(source))
	}
	return strings.ToLower(strings.TrimSpace(repo.URL))
}

func (s *StarRepoStorage) saveLocalStateLocked(repos []sourcedomain.StarRepo) error {
	snapshot := starReposLocalSnapshot{Repos: map[string]localRepoState{}}
	for _, repo := range repos {
		if repo.LastSync.IsZero() && strings.TrimSpace(repo.SyncError) == "" {
			continue
		}
		key := s.localStateKey(repo)
		if key == "" {
			continue
		}
		snapshot.Repos[key] = localRepoState{
			LastSync:  repo.LastSync,
			SyncError: repo.SyncError,
		}
	}
	if len(snapshot.Repos) == 0 {
		if err := os.Remove(s.localPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	}
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(s.localPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".star_repos_local_*.json")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		tmp.Close()
		os.Remove(tmpName)
	}()
	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, s.localPath)
}

func (s *StarRepoStorage) loadLocalStateLocked() (map[string]localRepoState, error) {
	data, err := os.ReadFile(s.localPath)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]localRepoState{}, nil
	}
	if err != nil {
		return nil, err
	}
	var snapshot starReposLocalSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, err
	}
	if snapshot.Repos == nil {
		return map[string]localRepoState{}, nil
	}
	return snapshot.Repos, nil
}

func (s *StarRepoStorage) resolveLocalDir(repo *sourcedomain.StarRepo) bool {
	resolved, needsMigration := pathutil.ResolveStoredPath(s.dataDir, repo.LocalDir, s.derivedLocalDir(repo.URL))
	repo.LocalDir = resolved
	return needsMigration
}

func (s *StarRepoStorage) derivedLocalDir(repoURL string) string {
	dir, err := platformgit.CacheDir(s.dataDir, repoURL)
	if err != nil {
		return ""
	}
	return dir
}

func (s *StarRepoStorage) buildBuiltinReposLocked() ([]sourcedomain.StarRepo, error) {
	repos := make([]sourcedomain.StarRepo, 0, len(s.builtinRepoURLs))
	for _, repoURL := range s.builtinRepoURLs {
		name, err := platformgit.ParseRepoName(repoURL)
		if err != nil {
			return nil, err
		}
		source, err := platformgit.RepoSource(repoURL)
		if err != nil {
			return nil, err
		}
		repos = append(repos, sourcedomain.StarRepo{
			URL:      repoURL,
			Name:     name,
			Source:   source,
			LocalDir: s.derivedLocalDir(repoURL),
		})
	}
	return repos, nil
}
