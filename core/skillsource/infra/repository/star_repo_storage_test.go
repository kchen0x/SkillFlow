package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	sourcedomain "github.com/shinerio/skillflow/core/skillsource/domain"
)

func TestStarRepoStorageLoadEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "star_repos.json")
	s := NewStarRepoStorage(path)
	repos, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if repos != nil {
		t.Fatalf("expected nil, got %v", repos)
	}
}

func TestStarRepoStorageSaveLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "star_repos.json")
	s := NewStarRepoStorage(path)
	localDir := filepath.Join(dir, "cache", "repos", "github.com", "a", "b")
	want := []sourcedomain.StarRepo{
		{URL: "https://github.com/a/b", Name: "a/b", LocalDir: localDir, LastSync: time.Time{}},
	}
	if err := s.Save(want); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), `"localDir"`) {
		t.Fatalf("expected synced file to exclude localDir, got %s", string(raw))
	}
	if strings.Contains(string(raw), `"lastSync"`) || strings.Contains(string(raw), `"syncError"`) {
		t.Fatalf("expected synced file to exclude local fields, got %s", string(raw))
	}
	got, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mismatch:\n got: %+v\nwant: %+v", got, want)
	}
}

func TestStarRepoStorageSaveLoadPersistsLocalStateInLocalFileOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "star_repos.json")
	s := NewStarRepoStorage(path)
	localDir := filepath.Join(dir, "cache", "repos", "github.com", "a", "b")
	lastSync := time.Now().UTC().Truncate(time.Second)
	want := []sourcedomain.StarRepo{
		{
			URL:       "https://github.com/a/b",
			Name:      "a/b",
			Source:    "github.com/a/b",
			LocalDir:  localDir,
			LastSync:  lastSync,
			SyncError: "network timeout",
		},
	}
	if err := s.Save(want); err != nil {
		t.Fatal(err)
	}

	syncedRaw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(syncedRaw), `"lastSync"`) || strings.Contains(string(syncedRaw), `"syncError"`) {
		t.Fatalf("expected synced file to exclude local state fields, got %s", string(syncedRaw))
	}

	localRaw, err := os.ReadFile(filepath.Join(dir, "star_repos_local.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(localRaw), `"lastSync"`) || !strings.Contains(string(localRaw), `"syncError"`) {
		t.Fatalf("expected local state file to include local fields, got %s", string(localRaw))
	}

	got, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(got))
	}
	if got[0].SyncError != "network timeout" {
		t.Fatalf("unexpected syncError: %q", got[0].SyncError)
	}
	if got[0].LastSync.UTC().Truncate(time.Second) != lastSync {
		t.Fatalf("unexpected lastSync: got %s want %s", got[0].LastSync.UTC().Format(time.RFC3339), lastSync.Format(time.RFC3339))
	}
}

func TestStarRepoStorageLoadMigratesLegacyLocalFieldsToLocalFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "star_repos.json")
	legacy := []sourcedomain.StarRepo{{
		URL:       "https://github.com/a/b",
		Name:      "a/b",
		Source:    "github.com/a/b",
		LocalDir:  "cache/repos/github.com/a/b",
		LastSync:  time.Now().UTC().Truncate(time.Second),
		SyncError: "legacy err",
	}}
	data, err := json.MarshalIndent(legacy, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	s := NewStarRepoStorage(path)
	got, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 repo, got %+v", got)
	}
	if got[0].SyncError != "legacy err" || got[0].LastSync.IsZero() {
		t.Fatalf("expected local state to survive migration, got %+v", got[0])
	}

	syncedRaw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(syncedRaw), `"lastSync"`) || strings.Contains(string(syncedRaw), `"syncError"`) || strings.Contains(string(syncedRaw), `"localDir"`) {
		t.Fatalf("expected migrated synced file to drop local-only fields, got %s", string(syncedRaw))
	}
	if _, err := os.Stat(filepath.Join(dir, "star_repos_local.json")); err != nil {
		t.Fatalf("expected local state file created after migration: %v", err)
	}
}

func TestStarRepoStorageLoadCorrupt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "star_repos.json")
	if err := os.WriteFile(path, []byte("{not valid json"), 0644); err != nil {
		t.Fatal(err)
	}
	s := NewStarRepoStorage(path)
	_, err := s.Load()
	if err == nil {
		t.Fatal("expected error for corrupt JSON, got nil")
	}
}

func TestStarRepoStorageLoadMigratesAbsoluteLocalDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "star_repos.json")
	localDir := filepath.Join(dir, "cache", "repos", "github.com", "a", "b")
	repos := []sourcedomain.StarRepo{{URL: "https://github.com/a/b", Name: "a/b", LocalDir: localDir}}
	data, err := json.MarshalIndent(repos, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	s := NewStarRepoStorage(path)
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
	if strings.Contains(string(raw), `"localDir"`) {
		t.Fatalf("expected migrated synced file to drop localDir, got %s", string(raw))
	}
}

func TestStarRepoStorageResolvesLocalDirFromCustomRepoCacheRoot(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "star_repos.json")
	cacheRoot := filepath.Join(dir, "volumes", "repo-cache")
	s := NewStarRepoStorageWithCacheDir(path, cacheRoot)
	repos := []sourcedomain.StarRepo{{
		URL:    "https://github.com/a/b",
		Name:   "a/b",
		Source: "github.com/a/b",
	}}
	if err := s.Save(repos); err != nil {
		t.Fatal(err)
	}

	got, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 repo, got %+v", got)
	}
	wantLocalDir := filepath.Join(cacheRoot, "github.com", "a", "b")
	if got[0].LocalDir != wantLocalDir {
		t.Fatalf("unexpected localDir: got %q want %q", got[0].LocalDir, wantLocalDir)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), `"localDir"`) {
		t.Fatalf("expected synced file to exclude localDir, got %s", string(raw))
	}
}

func TestStarRepoStorageLoadSeedsBuiltinsOnlyOnFirstInit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "star_repos.json")
	builtins := []string{
		"https://github.com/anthropics/skills.git",
		"https://github.com/ComposioHQ/awesome-claude-skills.git",
		"https://github.com/affaan-m/everything-claude-code.git",
	}
	s := NewStarRepoStorageWithBuiltins(path, builtins)

	first, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != len(builtins) {
		t.Fatalf("expected %d builtin repos, got %d", len(builtins), len(first))
	}
	for _, wantURL := range builtins {
		found := false
		for _, gotRepo := range first {
			if gotRepo.URL == wantURL {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected seeded repo list to include %s, got %+v", wantURL, first)
		}
	}

	if err := s.Save([]sourcedomain.StarRepo{}); err != nil {
		t.Fatal(err)
	}

	second, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(second) != 0 {
		t.Fatalf("expected empty repos after user deletion, got %d", len(second))
	}
}

func TestStarRepoStorageLoadDoesNotSeedWhenBuiltinsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "star_repos.json")
	s := NewStarRepoStorageWithBuiltins(path, nil)
	repos, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if repos != nil {
		t.Fatalf("expected nil repos without builtins, got %+v", repos)
	}
}
