package backup

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
)

type Snapshot map[string]SnapshotEntry

type SnapshotEntry struct {
	Size int64  `json:"size"`
	Hash string `json:"hash"`
}

func BuildSnapshot(root string) (Snapshot, error) {
	snapshot := make(Snapshot)

	if _, err := os.Stat(root); errors.Is(err, os.ErrNotExist) {
		return snapshot, nil
	} else if err != nil {
		return nil, err
	}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		if ShouldSkipBackupPath(rel) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			return nil
		}

		hash, err := hashFile(path)
		if err != nil {
			return err
		}
		snapshot[rel] = SnapshotEntry{
			Size: info.Size(),
			Hash: hash,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return snapshot, nil
}

func DiffSnapshots(previous, current Snapshot) []RemoteFile {
	changes := make([]RemoteFile, 0)

	for path, entry := range current {
		prev, ok := previous[path]
		switch {
		case !ok:
			changes = append(changes, RemoteFile{Path: path, Size: entry.Size, Action: "added"})
		case prev.Hash != entry.Hash:
			changes = append(changes, RemoteFile{Path: path, Size: entry.Size, Action: "modified"})
		}
	}

	for path, entry := range previous {
		if _, ok := current[path]; ok {
			continue
		}
		changes = append(changes, RemoteFile{Path: path, Size: entry.Size, Action: "deleted"})
	}

	sort.Slice(changes, func(i, j int) bool {
		if changes[i].Path == changes[j].Path {
			return changes[i].Action < changes[j].Action
		}
		return changes[i].Path < changes[j].Path
	})

	return changes
}

func LoadSnapshot(path string) (Snapshot, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return make(Snapshot), nil
	}
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return make(Snapshot), nil
	}

	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, err
	}
	if snapshot == nil {
		snapshot = make(Snapshot)
	}
	return snapshot, nil
}

func SaveSnapshot(path string, snapshot Snapshot) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
