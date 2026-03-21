package domain_test

import (
	"testing"

	"github.com/shinerio/skillflow/core/agentintegration/domain"
	"github.com/stretchr/testify/assert"
)

func TestAgentPresenceIndexAddsAgentsOncePerKey(t *testing.T) {
	presence := domain.NewAgentPresenceIndex()

	presence.Add("codex", "content:alpha", "content:alpha", "")
	presence.Add("claude-code", "content:alpha")
	presence.Add("codex", "content:beta")

	assert.Equal(t, []string{"claude-code", "codex"}, presence.Agents("content:alpha"))
	assert.Equal(t, []string{"codex"}, presence.Agents("content:beta"))
}
