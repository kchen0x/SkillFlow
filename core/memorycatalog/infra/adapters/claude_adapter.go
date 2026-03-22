package adapters

import (
	"github.com/shinerio/skillflow/core/memorycatalog/app/port/gateway"
	"github.com/shinerio/skillflow/core/memorycatalog/domain"
)

// ClaudeCodeAdapter implements AgentMemoryPusher for Claude Code.
// Claude Code auto-discovers rules files so BuildRulesIndex returns empty.
type ClaudeCodeAdapter struct{ *baseAdapter }

// NewClaudeCodeAdapter returns a new ClaudeCodeAdapter.
func NewClaudeCodeAdapter() *ClaudeCodeAdapter {
	return &ClaudeCodeAdapter{&baseAdapter{}}
}

// BuildRulesIndex returns an empty RulesIndex because Claude Code auto-scans its rules directory.
func (a *ClaudeCodeAdapter) BuildRulesIndex(_ []*domain.ModuleMemory, _ string) gateway.RulesIndex {
	return gateway.RulesIndex{}
}
