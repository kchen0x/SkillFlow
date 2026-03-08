package backup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiffSnapshots(t *testing.T) {
	previous := Snapshot{
		"a.txt": {Size: 1, Hash: "old-a"},
		"b.txt": {Size: 1, Hash: "same"},
		"c.txt": {Size: 1, Hash: "gone"},
	}
	current := Snapshot{
		"a.txt": {Size: 2, Hash: "new-a"},
		"b.txt": {Size: 1, Hash: "same"},
		"d.txt": {Size: 3, Hash: "new-d"},
	}

	changes := DiffSnapshots(previous, current)
	if len(changes) != 3 {
		t.Fatalf("expected 3 changes, got %d", len(changes))
	}

	expected := []RemoteFile{
		{Path: "a.txt", Size: 2, Action: "modified"},
		{Path: "c.txt", Size: 1, Action: "deleted"},
		{Path: "d.txt", Size: 3, Action: "added"},
	}
	for i, want := range expected {
		if changes[i] != want {
			t.Fatalf("change %d mismatch: got %+v want %+v", i, changes[i], want)
		}
	}
}

func TestBuildSnapshotSkipsExcludedPaths(t *testing.T) {
	root := t.TempDir()

	mustWriteFile(t, filepath.Join(root, "skills", "demo", "skill.md"), "demo")
	mustWriteFile(t, filepath.Join(root, "cache", "temp.txt"), "skip")
	mustWriteFile(t, filepath.Join(root, "config_local.json"), "skip")

	snapshot, err := BuildSnapshot(root)
	if err != nil {
		t.Fatalf("BuildSnapshot returned error: %v", err)
	}

	if _, ok := snapshot["skills/demo/skill.md"]; !ok {
		t.Fatal("expected skills/demo/skill.md in snapshot")
	}
	if _, ok := snapshot["cache/temp.txt"]; ok {
		t.Fatal("did not expect cache/temp.txt in snapshot")
	}
	if _, ok := snapshot["config_local.json"]; ok {
		t.Fatal("did not expect config_local.json in snapshot")
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
}
