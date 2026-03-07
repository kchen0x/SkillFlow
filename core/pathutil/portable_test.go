package pathutil

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestStorePathUsesPortableRelativePath(t *testing.T) {
	base := t.TempDir()
	current := filepath.Join(base, "skills", "coding", "skill-a")
	got := StorePath(base, current, "")
	if got != "skills/coding/skill-a" {
		t.Fatalf("StorePath()=%q, want %q", got, "skills/coding/skill-a")
	}
}

func TestResolveStoredPathFallsBackFromForeignAbsolute(t *testing.T) {
	base := t.TempDir()
	fallback := filepath.Join(base, "skills", "coding", "skill-a")
	foreign := `C:\Users\demo\.skillflow\skills\coding\skill-a`
	if runtime.GOOS == "windows" {
		foreign = "/Users/demo/.skillflow/skills/coding/skill-a"
	}
	got, migrated := ResolveStoredPath(base, foreign, fallback)
	if got != fallback {
		t.Fatalf("ResolveStoredPath() path=%q, want %q", got, fallback)
	}
	if !migrated {
		t.Fatal("ResolveStoredPath() should request migration for foreign absolute paths")
	}
}
