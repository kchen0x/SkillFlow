package domain

import (
	"os"
	"path/filepath"
)

var builtinAgentNames = []string{"claude-code", "opencode", "codex", "gemini-cli", "openclaw"}

func BuiltinAgentNames() []string {
	return append([]string(nil), builtinAgentNames...)
}

func DefaultProfile(agentName string) AgentProfile {
	profile := AgentProfile{
		Name:    agentName,
		Enabled: true,
		Custom:  false,
	}
	profile.ScanDirs = defaultScanDirs(agentName)
	if len(profile.ScanDirs) > 0 {
		profile.PushDir = profile.ScanDirs[0]
	}
	return profile
}

func defaultScanDirs(agentName string) []string {
	home, _ := os.UserHomeDir()
	agentsDir := filepath.Join(home, ".agents", "skills")

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
	return append([]string(nil), dirs[agentName]...)
}
