package adapters

import (
	"github.com/shinerio/skillflow/core/memorycatalog/app/port/gateway"
	"github.com/shinerio/skillflow/core/memorycatalog/domain"
)

// CodexAdapter implements AgentMemoryPusher for OpenAI Codex.
type CodexAdapter struct{ *baseAdapter }

// NewCodexAdapter returns a new CodexAdapter.
func NewCodexAdapter() *CodexAdapter {
	return &CodexAdapter{&baseAdapter{}}
}

// BuildRulesIndex returns explicit markdown refs for managed rule files.
func (a *CodexAdapter) BuildRulesIndex(modules []*domain.ModuleMemory, agentMemoryPath string, agentRulesDir string) gateway.RulesIndex {
	return buildExplicitRulesIndex(modules, agentMemoryPath, agentRulesDir)
}
