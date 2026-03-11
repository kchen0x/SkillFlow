package git

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/shinerio/skillflow/core/pathutil"
)

type StarStorage struct {
	path            string
	localPath       string
	dataDir         string
	builtinRepoURLs []string
	mu              sync.Mutex
}

func NewStarStorage(path string) *StarStorage {
	return NewStarStorageWithBuiltins(path, nil)
}

func NewStarStorageWithBuiltins(path string, builtinRepoURLs []string) *StarStorage {
	cleanPath := filepath.Clean(path)
	builtins := append([]string(nil), builtinRepoURLs...)
	dataDir := filepath.Dir(cleanPath)
	return &StarStorage{
		path:            cleanPath,
		localPath:       filepath.Join(dataDir, "star_repos_local.json"),
		dataDir:         dataDir,
		builtinRepoURLs: builtins,
	}
}

func (s *StarStorage) Load() ([]StarredRepo, error) {
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
	var repos []StarredRepo
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

func (s *StarStorage) Save(repos []StarredRepo) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked(repos)
}

func (s *StarStorage) saveLocked(repos []StarredRepo) error {
	if repos == nil {
		repos = []StarredRepo{}
	}
	snapshot := make([]syncedStarredRepo, len(repos))
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

type syncedStarredRepo struct {
	URL      string `json:"url"`
	Name     string `json:"name"`
	Source   string `json:"source"`
	LocalDir string `json:"localDir"`
}

func (s *StarStorage) serializedRepo(repo StarredRepo) syncedStarredRepo {
	return syncedStarredRepo{
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

type starredReposLocalSnapshot struct {
	Repos map[string]localRepoState `json:"repos"`
}

func (s *StarStorage) localStateKey(repo StarredRepo) string {
	if source := strings.TrimSpace(repo.Source); source != "" {
		return strings.ToLower(source)
	}
	if source, err := RepoSource(repo.URL); err == nil && strings.TrimSpace(source) != "" {
		return strings.ToLower(strings.TrimSpace(source))
	}
	return strings.ToLower(strings.TrimSpace(repo.URL))
}

func (s *StarStorage) saveLocalStateLocked(repos []StarredRepo) error {
	snapshot := starredReposLocalSnapshot{Repos: map[string]localRepoState{}}
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

func (s *StarStorage) loadLocalStateLocked() (map[string]localRepoState, error) {
	data, err := os.ReadFile(s.localPath)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]localRepoState{}, nil
	}
	if err != nil {
		return nil, err
	}
	var snapshot starredReposLocalSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, err
	}
	if snapshot.Repos == nil {
		return map[string]localRepoState{}, nil
	}
	return snapshot.Repos, nil
}

func (s *StarStorage) resolveLocalDir(repo *StarredRepo) bool {
	resolved, needsMigration := pathutil.ResolveStoredPath(s.dataDir, repo.LocalDir, s.derivedLocalDir(repo.URL))
	repo.LocalDir = resolved
	return needsMigration
}

func (s *StarStorage) derivedLocalDir(repoURL string) string {
	dir, err := CacheDir(s.dataDir, repoURL)
	if err != nil {
		return ""
	}
	return dir
}

func (s *StarStorage) buildBuiltinReposLocked() ([]StarredRepo, error) {
	repos := make([]StarredRepo, 0, len(s.builtinRepoURLs))
	for _, repoURL := range s.builtinRepoURLs {
		name, err := ParseRepoName(repoURL)
		if err != nil {
			return nil, err
		}
		source, err := RepoSource(repoURL)
		if err != nil {
			return nil, err
		}
		repos = append(repos, StarredRepo{
			URL:      repoURL,
			Name:     name,
			Source:   source,
			LocalDir: s.derivedLocalDir(repoURL),
		})
	}
	return repos, nil
}
