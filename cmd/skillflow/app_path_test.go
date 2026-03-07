package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNearestExistingDirectoryReturnsDirectory(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "child")
	if err := os.MkdirAll(child, 0755); err != nil {
		t.Fatal(err)
	}
	if got := nearestExistingDirectory(child); got != child {
		t.Fatalf("nearestExistingDirectory()=%q, want %q", got, child)
	}
}

func TestNearestExistingDirectoryFallsBackToExistingParent(t *testing.T) {
	root := t.TempDir()
	missing := filepath.Join(root, "missing", "deep")
	if got := nearestExistingDirectory(missing); got != root {
		t.Fatalf("nearestExistingDirectory()=%q, want %q", got, root)
	}
}

func TestResolveOpenPathTargetUsesParentForFiles(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "skill.md")
	if err := os.WriteFile(file, []byte("# skill"), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := resolveOpenPathTarget(file)
	if err != nil {
		t.Fatal(err)
	}
	if got != root {
		t.Fatalf("resolveOpenPathTarget()=%q, want %q", got, root)
	}
}
