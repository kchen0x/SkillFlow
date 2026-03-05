package git

import (
	"os"
	"path/filepath"
	"reflect"
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
	path := filepath.Join(t.TempDir(), "star_repos.json")
	s := NewStarStorage(path)
	want := []StarredRepo{
		{URL: "https://github.com/a/b", Name: "a/b", LocalDir: "/tmp/a/b", LastSync: time.Time{}},
	}
	if err := s.Save(want); err != nil {
		t.Fatal(err)
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
