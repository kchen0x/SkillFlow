package config

import "github.com/shinerio/skillflow/core/readmodel/preferences"

const (
	SkillStatusImported     = preferences.SkillStatusImported
	SkillStatusUpdatable    = preferences.SkillStatusUpdatable
	SkillStatusPushedAgents = preferences.SkillStatusPushedAgents
)

type SkillStatusVisibilityConfig = preferences.SkillStatusVisibilityConfig

func DefaultSkillStatusVisibility() SkillStatusVisibilityConfig {
	return preferences.DefaultSkillStatusVisibility()
}

func NormalizeSkillStatusVisibility(cfg SkillStatusVisibilityConfig) SkillStatusVisibilityConfig {
	return preferences.NormalizeSkillStatusVisibility(cfg)
}
