package snapshot

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"

	backupdomain "github.com/shinerio/skillflow/core/backup/domain"
)

func BuildSnapshot(root string) (backupdomain.Snapshot, error) {
	snapshot := make(backupdomain.Snapshot)

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
		snapshot[rel] = backupdomain.SnapshotEntry{
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

func DiffSnapshots(previous, current backupdomain.Snapshot) []backupdomain.RemoteFile {
	changes := make([]backupdomain.RemoteFile, 0)

	for path, entry := range current {
		prev, ok := previous[path]
		switch {
		case !ok:
			changes = append(changes, backupdomain.RemoteFile{Path: path, Size: entry.Size, Action: "added"})
		case prev.Hash != entry.Hash:
			changes = append(changes, backupdomain.RemoteFile{Path: path, Size: entry.Size, Action: "modified"})
		}
	}

	for path, entry := range previous {
		if _, ok := current[path]; ok {
			continue
		}
		changes = append(changes, backupdomain.RemoteFile{Path: path, Size: entry.Size, Action: "deleted"})
	}

	sort.Slice(changes, func(i, j int) bool {
		if changes[i].Path == changes[j].Path {
			return changes[i].Action < changes[j].Action
		}
		return changes[i].Path < changes[j].Path
	})

	return changes
}

func LoadSnapshot(path string) (backupdomain.Snapshot, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return make(backupdomain.Snapshot), nil
	}
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return make(backupdomain.Snapshot), nil
	}

	var snapshot backupdomain.Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, err
	}
	if snapshot == nil {
		snapshot = make(backupdomain.Snapshot)
	}
	return snapshot, nil
}

func SaveSnapshot(path string, snapshot backupdomain.Snapshot) error {
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
