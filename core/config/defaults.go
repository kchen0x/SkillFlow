package config

import (
	"os"
	"path/filepath"
	"runtime"

	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
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

func DefaultConfig(dataDir string) AppConfig {
	names := agentdomain.BuiltinAgentNames()
	agents := make([]AgentConfig, 0, len(names))
	for _, name := range names {
		profile := agentdomain.DefaultProfile(name)
		agents = append(agents, AgentConfig{
			Name:     profile.Name,
			ScanDirs: profile.ScanDirs,
			PushDir:  profile.PushDir,
			Enabled:  profile.Enabled,
			Custom:   profile.Custom,
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
