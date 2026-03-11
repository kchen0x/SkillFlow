package viewstate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRebuildToolPresenceRescansOnlyChangedTools(t *testing.T) {
	previous := ToolPresenceSnapshot{
		ToolFingerprints: map[string]string{
			"codex":  "fp-codex-1",
			"claude": "fp-claude-1",
		},
		KeysByTool: map[string][]string{
			"codex":  {"skill:a", "skill:b"},
			"claude": {"skill:c"},
		},
	}

	var scanned []string
	next, err := RebuildToolPresence(previous, []ToolPresenceInput{
		{Name: "codex", Fingerprint: "fp-codex-2"},
		{Name: "claude", Fingerprint: "fp-claude-1"},
	}, func(toolName string) ([]string, error) {
		scanned = append(scanned, toolName)
		if toolName == "codex" {
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
	}, next.ToolsByKey)
	assert.Equal(t, map[string][]string{
		"codex":  {"skill:a", "skill:d"},
		"claude": {"skill:c"},
	}, next.KeysByTool)
}

func TestRebuildToolPresenceDropsRemovedTools(t *testing.T) {
	previous := ToolPresenceSnapshot{
		ToolFingerprints: map[string]string{
			"codex":  "fp-codex-1",
			"claude": "fp-claude-1",
		},
		KeysByTool: map[string][]string{
			"codex":  {"skill:a"},
			"claude": {"skill:c"},
		},
	}

	next, err := RebuildToolPresence(previous, []ToolPresenceInput{
		{Name: "codex", Fingerprint: "fp-codex-1"},
	}, func(toolName string) ([]string, error) {
		return nil, nil
	})
	require.NoError(t, err)

	assert.Equal(t, map[string][]string{
		"skill:a": {"codex"},
	}, next.ToolsByKey)
	assert.Equal(t, map[string][]string{
		"codex": {"skill:a"},
	}, next.KeysByTool)
	assert.Equal(t, map[string]string{
		"codex": "fp-codex-1",
	}, next.ToolFingerprints)
}
