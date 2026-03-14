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
	require.Len(t, agents, 1)
	assert.Equal(t, "codex", agents[0].Name)
}

func TestPushConflictUsesAgentNameField(t *testing.T) {
	conflict := PushConflict{
		SkillName:  "demo-skill",
		AgentName:  "codex",
		TargetPath: "/tmp/codex/demo-skill",
	}

	assert.Equal(t, "codex", conflict.AgentName)
}
