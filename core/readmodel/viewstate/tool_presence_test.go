package viewstate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRebuildAgentPresenceRescansOnlyChangedAgents(t *testing.T) {
	previous := AgentPresenceSnapshot{
		AgentFingerprints: map[string]string{
			"codex":  "fp-codex-1",
			"claude": "fp-claude-1",
		},
		KeysByAgent: map[string][]string{
			"codex":  {"skill:a", "skill:b"},
			"claude": {"skill:c"},
		},
	}

	var scanned []string
	next, err := RebuildAgentPresence(previous, []AgentPresenceInput{
		{Name: "codex", Fingerprint: "fp-codex-2"},
		{Name: "claude", Fingerprint: "fp-claude-1"},
	}, func(agentName string) ([]string, error) {
		scanned = append(scanned, agentName)
		if agentName == "codex" {
			return []string{"skill:a", "skill:d"}, nil
		}
		return nil, nil
	})
	require.NoError(t, err)

	assert.Equal(t, []string{"codex"}, scanned)
	assert.Equal(t, map[string][]string{
		"skill:a": {"codex"},
		"skill:c": {"claude"},
		"skill:d": {"codex"},
	}, next.AgentsByKey)
	assert.Equal(t, map[string][]string{
		"codex":  {"skill:a", "skill:d"},
		"claude": {"skill:c"},
	}, next.KeysByAgent)
}

func TestRebuildAgentPresenceDropsRemovedAgents(t *testing.T) {
	previous := AgentPresenceSnapshot{
		AgentFingerprints: map[string]string{
			"codex":  "fp-codex-1",
			"claude": "fp-claude-1",
		},
		KeysByAgent: map[string][]string{
			"codex":  {"skill:a"},
			"claude": {"skill:c"},
		},
	}

	next, err := RebuildAgentPresence(previous, []AgentPresenceInput{
		{Name: "codex", Fingerprint: "fp-codex-1"},
	}, func(agentName string) ([]string, error) {
		return nil, nil
	})
	require.NoError(t, err)

	assert.Equal(t, map[string][]string{
		"skill:a": {"codex"},
	}, next.AgentsByKey)
	assert.Equal(t, map[string][]string{
		"codex": {"skill:a"},
	}, next.KeysByAgent)
	assert.Equal(t, map[string]string{
		"codex": "fp-codex-1",
	}, next.AgentFingerprints)
}
