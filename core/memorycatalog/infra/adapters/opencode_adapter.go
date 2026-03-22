package adapters

import (
	"github.com/shinerio/skillflow/core/memorycatalog/app/port/gateway"
	"github.com/shinerio/skillflow/core/memorycatalog/domain"
)

// OpenCodeAdapter implements AgentMemoryPusher for OpenCode.
type OpenCodeAdapter struct{ *baseAdapter }

// NewOpenCodeAdapter returns a new OpenCodeAdapter.
func NewOpenCodeAdapter() *OpenCodeAdapter {
	return &OpenCodeAdapter{&baseAdapter{}}
}

// BuildRulesIndex returns explicit markdown refs for managed rule files.
func (a *OpenCodeAdapter) BuildRulesIndex(modules []*domain.ModuleMemory, agentMemoryPath string, agentRulesDir string) gateway.RulesIndex {
	return buildExplicitRulesIndex(modules, agentMemoryPath, agentRulesDir)
}
