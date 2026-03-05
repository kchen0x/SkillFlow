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

func defaultAgentsSkillsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".agents", "skills")
}

func DefaultToolScanDirs(toolName string) []string {
	home, _ := os.UserHomeDir()
	agentsDir := defaultAgentsSkillsDir()

	dirs := map[string]map[string][]string{
		"darwin": {
			"claude-code": {
				filepath.Join(home, ".claude", "skills"),
				filepath.Join(home, ".claude", "plugins", "marketplaces"),
			},
			"opencode": {
				filepath.Join(home, ".config", "opencode", "skills"),
				agentsDir,
			},
			"codex": {
				agentsDir,
			},
			"gemini-cli": {
				filepath.Join(home, ".gemini", "skills"),
				agentsDir,
			},
			"openclaw": {
				filepath.Join(home, ".openclaw", "skills"),
				filepath.Join(home, ".openclaw", "workspace", "skills"),
			},
		},
		"windows": {
			"claude-code": {
				filepath.Join(os.Getenv("APPDATA"), "claude", "skills"),
				filepath.Join(os.Getenv("APPDATA"), "claude", "plugins", "marketplaces"),
			},
			"opencode": {
				filepath.Join(os.Getenv("APPDATA"), "opencode", "skills"),
				agentsDir,
			},
			"codex": {
				agentsDir,
			},
			"gemini-cli": {
				filepath.Join(os.Getenv("APPDATA"), "gemini", "skills"),
				agentsDir,
			},
			"openclaw": {
				filepath.Join(os.Getenv("APPDATA"), "openclaw", "skills"),
				filepath.Join(os.Getenv("APPDATA"), "openclaw", "workspace", "skills"),
			},
		},
	}
	goos := runtime.GOOS
	if goos != "windows" {
		goos = "darwin"
	}
	return dirs[goos][toolName]
}

// DefaultToolsDir returns the default push path for a tool.
func DefaultToolsDir(toolName string) string {
	scanDirs := DefaultToolScanDirs(toolName)
	if len(scanDirs) == 0 {
		return ""
	}
	return scanDirs[0]
}

var builtinTools = []string{"claude-code", "opencode", "codex", "gemini-cli", "openclaw"}

func DefaultConfig(dataDir string) AppConfig {
	tools := make([]ToolConfig, 0, len(builtinTools))
	for _, name := range builtinTools {
		scanDirs := DefaultToolScanDirs(name)
		pushDir := DefaultToolsDir(name)
		enabled := false
		for _, dir := range scanDirs {
			if _, err := os.Stat(dir); err == nil {
				enabled = true
				break
			}
		}

		tools = append(tools, ToolConfig{
			Name:     name,
			ScanDirs: scanDirs,
			PushDir:  pushDir,
			Enabled:  enabled,
			Custom:   false,
		})
	}
	return AppConfig{
		SkillsStorageDir: filepath.Join(dataDir, "skills"),
		DefaultCategory:  "Imported",
		Tools:            tools,
		Cloud:            CloudConfig{RemotePath: "skillflow/"},
	}
}
