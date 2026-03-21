package preferences

import "strings"

const (
	SkillStatusImported     = "imported"
	SkillStatusUpdatable    = "updatable"
	SkillStatusPushedAgents = "pushedAgents"
)

var allSkillStatusKeys = []string{
	SkillStatusImported,
	SkillStatusUpdatable,
	SkillStatusPushedAgents,
}

type SkillStatusVisibilityConfig struct {
	MySkills      []string `json:"mySkills"`
	MyAgents      []string `json:"myAgents"`
	PushToAgent   []string `json:"pushToAgent"`
	PullFromAgent []string `json:"pullFromAgent"`
	StarredRepos  []string `json:"starredRepos"`
}

func DefaultSkillStatusVisibility() SkillStatusVisibilityConfig {
	return SkillStatusVisibilityConfig{
		MySkills:      []string{SkillStatusUpdatable, SkillStatusPushedAgents},
		MyAgents:      []string{SkillStatusImported, SkillStatusUpdatable, SkillStatusPushedAgents},
		PushToAgent:   []string{SkillStatusPushedAgents},
		PullFromAgent: []string{SkillStatusImported},
		StarredRepos:  []string{SkillStatusImported, SkillStatusPushedAgents},
	}
}

func NormalizeSkillStatusVisibility(cfg SkillStatusVisibilityConfig) SkillStatusVisibilityConfig {
	defaults := DefaultSkillStatusVisibility()
	return SkillStatusVisibilityConfig{
		MySkills:      normalizeSkillStatusKeys(cfg.MySkills, defaults.MySkills),
		MyAgents:      normalizeSkillStatusKeys(cfg.MyAgents, defaults.MyAgents),
		PushToAgent:   normalizeSkillStatusKeys(cfg.PushToAgent, defaults.PushToAgent),
		PullFromAgent: normalizeSkillStatusKeys(cfg.PullFromAgent, defaults.PullFromAgent),
		StarredRepos:  normalizeSkillStatusKeys(cfg.StarredRepos, defaults.StarredRepos),
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
