package git

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestStarStorageLoadEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "star_repos.json")
	s := NewStarStorage(path)
	repos, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if repos != nil {
		t.Fatalf("expected nil, got %v", repos)
	}
}

func TestStarStorageSaveLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "star_repos.json")
	s := NewStarStorage(path)
	localDir, err := CacheDir(dir, "https://github.com/a/b")
	if err != nil {
		t.Fatal(err)
	}
	want := []StarredRepo{
		{URL: "https://github.com/a/b", Name: "a/b", LocalDir: localDir, LastSync: time.Time{}},
	}
	if err := s.Save(want); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"localDir": "cache/github.com/a/b"`) {
		t.Fatalf("expected relative localDir in persisted file, got %s", string(raw))
	}
	got, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mismatch:\n got: %+v\nwant: %+v", got, want)
	}
}

func TestStarStorageLoadCorrupt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "star_repos.json")
	if err := os.WriteFile(path, []byte("{not valid json"), 0644); err != nil {
		t.Fatal(err)
	}
	s := NewStarStorage(path)
	_, err := s.Load()
	if err == nil {
		t.Fatal("expected error for corrupt JSON, got nil")
	}
}

func TestStarStorageLoadMigratesAbsoluteLocalDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "star_repos.json")
	localDir, err := CacheDir(dir, "https://github.com/a/b")
	if err != nil {
		t.Fatal(err)
	}
	repos := []StarredRepo{{URL: "https://github.com/a/b", Name: "a/b", LocalDir: localDir}}
	data, err := json.MarshalIndent(repos, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	s := NewStarStorage(path)
	got, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].LocalDir != localDir {
		t.Fatalf("unexpected load result: %+v", got)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"localDir": "cache/github.com/a/b"`) {
		t.Fatalf("expected migrated relative localDir, got %s", string(raw))
	}
}
