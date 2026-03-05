package git

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type StarStorage struct {
	path string
	mu   sync.Mutex
}

func NewStarStorage(path string) *StarStorage {
	return &StarStorage{path: path}
}

func (s *StarStorage) Load() ([]StarredRepo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var repos []StarredRepo
	return repos, json.Unmarshal(data, &repos)
}

func (s *StarStorage) Save(repos []StarredRepo) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if repos == nil {
		repos = []StarredRepo{}
	}
	data, err := json.MarshalIndent(repos, "", "  ")
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
	return os.Rename(tmpName, s.path)
}
