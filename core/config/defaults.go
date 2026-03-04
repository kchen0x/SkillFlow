package config

import (
	"os"
	"path/filepath"
	"runtime"
)

func AppDataDir() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "SkillFlow")
	default: // darwin
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "SkillFlow")
	}
}

func DefaultToolsDir(toolName string) string {
	home, _ := os.UserHomeDir()
	dirs := map[string]map[string]string{
		"darwin": {
			"claude-code": filepath.Join(home, ".claude", "skills"),
			"opencode":    filepath.Join(home, ".opencode", "skills"),
			"codex":       filepath.Join(home, ".codex", "skills"),
			"gemini-cli":  filepath.Join(home, ".gemini", "skills"),
			"openclaw":    filepath.Join(home, ".openclaw", "skills"),
		},
		"windows": {
			"claude-code": filepath.Join(os.Getenv("APPDATA"), "claude", "skills"),
			"opencode":    filepath.Join(os.Getenv("APPDATA"), "opencode", "skills"),
			"codex":       filepath.Join(os.Getenv("APPDATA"), "codex", "skills"),
			"gemini-cli":  filepath.Join(os.Getenv("APPDATA"), "gemini", "skills"),
			"openclaw":    filepath.Join(os.Getenv("APPDATA"), "openclaw", "skills"),
		},
	}
	goos := runtime.GOOS
	if goos != "windows" {
		goos = "darwin"
	}
	return dirs[goos][toolName]
}

var builtinTools = []string{"claude-code", "opencode", "codex", "gemini-cli", "openclaw"}

func DefaultConfig(dataDir string) AppConfig {
	tools := make([]ToolConfig, 0, len(builtinTools))
	for _, name := range builtinTools {
		dir := DefaultToolsDir(name)
		_, err := os.Stat(dir)
		tools = append(tools, ToolConfig{
			Name:      name,
			SkillsDir: dir,
			Enabled:   err == nil,
			Custom:    false,
		})
	}
	return AppConfig{
		SkillsStorageDir: filepath.Join(dataDir, "skills"),
		DefaultCategory:  "Imported",
		Tools:            tools,
		Cloud:            CloudConfig{RemotePath: "skillflow/"},
	}
}
