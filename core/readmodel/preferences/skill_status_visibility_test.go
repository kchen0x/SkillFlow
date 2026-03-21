package preferences_test

import (
	"testing"

	"github.com/shinerio/skillflow/core/readmodel/preferences"
	"github.com/stretchr/testify/assert"
)

func TestDefaultSkillStatusVisibility(t *testing.T) {
	defaults := preferences.DefaultSkillStatusVisibility()
	assert.Equal(t, []string{preferences.SkillStatusUpdatable, preferences.SkillStatusPushedAgents}, defaults.MySkills)
	assert.Equal(t, []string{preferences.SkillStatusImported}, defaults.PullFromAgent)
}

func TestNormalizeSkillStatusVisibilityDropsKeysOutsidePagePolicy(t *testing.T) {
	normalized := preferences.NormalizeSkillStatusVisibility(preferences.SkillStatusVisibilityConfig{
		PullFromAgent: []string{preferences.SkillStatusImported, preferences.SkillStatusPushedAgents},
		PushToAgent:   []string{preferences.SkillStatusImported, preferences.SkillStatusPushedAgents},
	})
	assert.Equal(t, []string{preferences.SkillStatusImported}, normalized.PullFromAgent)
	assert.Equal(t, []string{preferences.SkillStatusPushedAgents}, normalized.PushToAgent)
}
