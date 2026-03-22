package adapters

import (
	"github.com/shinerio/skillflow/core/memorycatalog/app/port/gateway"
	"github.com/shinerio/skillflow/core/memorycatalog/domain"
)

// CustomAdapter implements AgentMemoryPusher for user-defined custom agents.
type CustomAdapter struct{ *baseAdapter }

// NewCustomAdapter returns a new CustomAdapter.
func NewCustomAdapter() *CustomAdapter {
	return &CustomAdapter{&baseAdapter{}}
}

// BuildRulesIndex returns an explicit listing of managed rule files.
func (a *CustomAdapter) BuildRulesIndex(modules []*domain.ModuleMemory, agentRulesDir string) gateway.RulesIndex {
	return buildExplicitRulesIndex(modules, agentRulesDir)
}
