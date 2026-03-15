package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEnabledAgentsReturnsEnabledAgents(t *testing.T) {
	app, _, _ := newAutoPushTestApp(t, []string{"codex"})

	agents, err := app.GetEnabledAgents()
	require.NoError(t, err)
	require.NotEmpty(t, agents)

	enabledNames := make([]string, 0, len(agents))
	for _, agent := range agents {
		assert.True(t, agent.Enabled)
		enabledNames = append(enabledNames, agent.Name)
	}
	assert.Contains(t, enabledNames, "codex")
}

func TestPushConflictUsesAgentNameField(t *testing.T) {
	conflict := PushConflict{
		SkillName:  "demo-skill",
		AgentName:  "codex",
		TargetPath: "/tmp/codex/demo-skill",
	}

	assert.Equal(t, "codex", conflict.AgentName)
}
