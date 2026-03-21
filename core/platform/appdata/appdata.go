package appdata

import (
	"os"
	"path/filepath"
	"runtime"
)

func Dir() string {
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(home, ".skillflow")
	default: // darwin / linux
		return filepath.Join(home, "Library", "Application Support", "SkillFlow")
	}
}

func SkillsDir(dataDir string) string {
	return filepath.Join(dataDir, "skills")
}

func RepoCacheDir(dataDir string) string {
	return filepath.Join(dataDir, "cache", "repos")
}
