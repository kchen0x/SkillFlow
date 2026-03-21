package settingsstore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Store struct {
	dataDir    string
	sharedPath string
	localPath  string
}

func New(dataDir string) *Store {
	return &Store{
		dataDir:    dataDir,
		sharedPath: filepath.Join(dataDir, "config.json"),
		localPath:  filepath.Join(dataDir, "config_local.json"),
	}
}

func (s *Store) DataDir() string {
	return s.dataDir
}

func (s *Store) SharedPath() string {
	return s.sharedPath
}

func (s *Store) LocalPath() string {
	return s.localPath
}

func (s *Store) ReadShared(out any) (bool, error) {
	return readJSON(s.sharedPath, out)
}

func (s *Store) ReadLocal(out any) (bool, error) {
	return readJSON(s.localPath, out)
}

func (s *Store) WriteShared(value any) error {
	return writeJSON(s.dataDir, s.sharedPath, value)
}

func (s *Store) WriteLocal(value any) error {
	return writeJSON(s.dataDir, s.localPath, value)
}

func readJSON(path string, out any) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if err := json.Unmarshal(data, out); err != nil {
		return false, err
	}
	return true, nil
}

func writeJSON(root string, path string, value any) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(root, strings.TrimSuffix(filepath.Base(path), ".json")+".*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	_ = os.Remove(path)
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}
