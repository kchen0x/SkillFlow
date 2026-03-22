package adapters

import (
	"github.com/shinerio/skillflow/core/memorycatalog/app/port/gateway"
	"github.com/shinerio/skillflow/core/memorycatalog/domain"
)

// OpenClawAdapter implements AgentMemoryPusher for OpenClaw.
type OpenClawAdapter struct{ *baseAdapter }

// NewOpenClawAdapter returns a new OpenClawAdapter.
func NewOpenClawAdapter() *OpenClawAdapter {
	return &OpenClawAdapter{&baseAdapter{}}
}

// BuildRulesIndex returns explicit markdown refs for managed rule files.
func (a *OpenClawAdapter) BuildRulesIndex(modules []*domain.ModuleMemory, agentRulesDir string) gateway.RulesIndex {
	return buildExplicitRulesIndex(modules, agentRulesDir)
}
