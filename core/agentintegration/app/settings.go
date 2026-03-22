package app

import (
	"strings"

	"github.com/shinerio/skillflow/core/agentintegration/domain"
)

const (
	DefaultRepoScanMaxDepth = 5
	MinRepoScanMaxDepth     = 1
	MaxRepoScanMaxDepth     = 20
)

type SharedAgentSettings struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

type SharedSettings struct {
	RepoScanMaxDepth int                   `json:"repoScanMaxDepth"`
	Agents           []SharedAgentSettings `json:"agents"`
}

type LocalAgentSettings struct {
	Name       string   `json:"name"`
	ScanDirs   []string `json:"scanDirs"`
	PushDir    string   `json:"pushDir"`
	MemoryPath string   `json:"memoryPath"`
	RulesDir   string   `json:"rulesDir"`
	Custom     bool     `json:"custom"`
	Enabled    bool     `json:"enabled"`
}

type LocalSettings struct {
	AutoPushAgents []string             `json:"autoPushAgents"`
	Agents         []LocalAgentSettings `json:"agents"`
}

type Settings struct {
	Shared SharedSettings
	Local  LocalSettings
}

func DefaultProfiles() []domain.AgentProfile {
	names := domain.BuiltinAgentNames()
	profiles := make([]domain.AgentProfile, 0, len(names))
	for _, name := range names {
		profiles = append(profiles, domain.DefaultProfile(name))
	}
	return profiles
}

func DefaultSharedSettings() SharedSettings {
	profiles := DefaultProfiles()
	agents := make([]SharedAgentSettings, 0, len(profiles))
	for _, profile := range profiles {
		agents = append(agents, SharedAgentSettings{
			Name:    profile.Name,
			Enabled: profile.Enabled,
		})
	}
	return SharedSettings{
		RepoScanMaxDepth: DefaultRepoScanMaxDepth,
		Agents:           agents,
	}
}

func DefaultLocalSettings() LocalSettings {
	profiles := DefaultProfiles()
	agents := make([]LocalAgentSettings, 0, len(profiles))
	for _, profile := range profiles {
		agents = append(agents, LocalAgentSettings{
			Name:       profile.Name,
			ScanDirs:   append([]string(nil), profile.ScanDirs...),
			PushDir:    profile.PushDir,
			MemoryPath: profile.MemoryPath,
			RulesDir:   profile.RulesDir,
		})
	}
	return LocalSettings{
		Agents: agents,
	}
}

func DefaultSettings() Settings {
	return Settings{
		Shared: DefaultSharedSettings(),
		Local:  DefaultLocalSettings(),
	}
}

func NormalizeAutoPushAgentNames(names []string) []string {
	if len(names) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(names))
	normalized := make([]string, 0, len(names))
	for _, name := range names {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func NormalizeRepoScanMaxDepth(depth int) int {
	if depth < MinRepoScanMaxDepth {
		return DefaultRepoScanMaxDepth
	}
	if depth > MaxRepoScanMaxDepth {
		return MaxRepoScanMaxDepth
	}
	return depth
}
