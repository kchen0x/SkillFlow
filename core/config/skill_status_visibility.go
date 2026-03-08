package config

import "strings"

const (
	SkillStatusImported    = "imported"
	SkillStatusUpdatable   = "updatable"
	SkillStatusPushedTools = "pushedTools"
)

var allSkillStatusKeys = []string{
	SkillStatusImported,
	SkillStatusUpdatable,
	SkillStatusPushedTools,
}

type SkillStatusVisibilityConfig struct {
	MySkills      []string `json:"mySkills"`
	MyTools       []string `json:"myTools"`
	PushToTool    []string `json:"pushToTool"`
	PullFromTool  []string `json:"pullFromTool"`
	StarredRepos  []string `json:"starredRepos"`
	GitHubInstall []string `json:"githubInstall"`
}

func DefaultSkillStatusVisibility() SkillStatusVisibilityConfig {
	return SkillStatusVisibilityConfig{
		MySkills:      []string{SkillStatusUpdatable, SkillStatusPushedTools},
		MyTools:       []string{SkillStatusImported, SkillStatusUpdatable, SkillStatusPushedTools},
		PushToTool:    []string{SkillStatusPushedTools},
		PullFromTool:  []string{SkillStatusImported},
		StarredRepos:  []string{SkillStatusImported, SkillStatusPushedTools},
		GitHubInstall: []string{SkillStatusImported, SkillStatusUpdatable, SkillStatusPushedTools},
	}
}

func NormalizeSkillStatusVisibility(cfg SkillStatusVisibilityConfig) SkillStatusVisibilityConfig {
	defaults := DefaultSkillStatusVisibility()
	return SkillStatusVisibilityConfig{
		MySkills:      normalizeSkillStatusKeys(cfg.MySkills, defaults.MySkills),
		MyTools:       normalizeSkillStatusKeys(cfg.MyTools, defaults.MyTools),
		PushToTool:    normalizeSkillStatusKeys(cfg.PushToTool, defaults.PushToTool),
		PullFromTool:  normalizeSkillStatusKeys(cfg.PullFromTool, defaults.PullFromTool),
		StarredRepos:  normalizeSkillStatusKeys(cfg.StarredRepos, defaults.StarredRepos),
		GitHubInstall: normalizeSkillStatusKeys(cfg.GitHubInstall, defaults.GitHubInstall),
	}
}

func normalizeSkillStatusKeys(current []string, defaults []string) []string {
	if current == nil {
		current = defaults
	}

	allowed := make(map[string]struct{}, len(defaults))
	for _, key := range defaults {
		allowed[key] = struct{}{}
	}

	enabled := make(map[string]struct{}, len(current))
	for _, key := range current {
		normalized := strings.TrimSpace(key)
		if normalized == "" {
			continue
		}
		if _, ok := allowed[normalized]; !ok {
			continue
		}
		enabled[normalized] = struct{}{}
	}

	out := make([]string, 0, len(allSkillStatusKeys))
	for _, key := range allSkillStatusKeys {
		if _, ok := enabled[key]; ok {
			out = append(out, key)
		}
	}
	return out
}
