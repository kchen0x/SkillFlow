package config

import (
	"os"
	"path/filepath"
	"runtime"
)

func AppDataDir() string {
	switch runtime.GOOS {
	case "windows":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".skillflow")
	default: // darwin / linux
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "SkillFlow")
	}
}

func defaultAgentsSkillsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".agents", "skills")
}

func DefaultAgentScanDirs(agentName string) []string {
	home, _ := os.UserHomeDir()
	agentsDir := defaultAgentsSkillsDir()

	// All platforms use home-relative paths.
	dirs := map[string][]string{
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
	}
	return dirs[agentName]
}

// DefaultAgentPushDir returns the default push path for an agent.
func DefaultAgentPushDir(agentName string) string {
	scanDirs := DefaultAgentScanDirs(agentName)
	if len(scanDirs) == 0 {
		return ""
	}
	return scanDirs[0]
}

var builtinAgents = []string{"claude-code", "opencode", "codex", "gemini-cli", "openclaw"}

func DefaultConfig(dataDir string) AppConfig {
	agents := make([]AgentConfig, 0, len(builtinAgents))
	for _, name := range builtinAgents {
		scanDirs := DefaultAgentScanDirs(name)
		pushDir := DefaultAgentPushDir(name)
		agents = append(agents, AgentConfig{
			Name:     name,
			ScanDirs: scanDirs,
			PushDir:  pushDir,
			Enabled:  true,
			Custom:   false,
		})
	}
	return AppConfig{
		SkillsStorageDir:      filepath.Join(dataDir, "skills"),
		DefaultCategory:       "Default",
		LogLevel:              DefaultLogLevel,
		RepoScanMaxDepth:      DefaultRepoScanMaxDepth,
		SkillStatusVisibility: DefaultSkillStatusVisibility(),
		Agents:                agents,
		Cloud:                 CloudConfig{RemotePath: DefaultCloudRemotePath},
		Proxy:                 ProxyConfig{Mode: ProxyModeNone},
	}
}
