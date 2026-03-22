package adapters

import (
	"github.com/shinerio/skillflow/core/memorycatalog/app/port/gateway"
	"github.com/shinerio/skillflow/core/memorycatalog/domain"
)

// GeminiAdapter implements AgentMemoryPusher for Google Gemini CLI.
type GeminiAdapter struct{ *baseAdapter }

// NewGeminiAdapter returns a new GeminiAdapter.
func NewGeminiAdapter() *GeminiAdapter {
	return &GeminiAdapter{&baseAdapter{}}
}

// BuildRulesIndex returns explicit markdown refs for managed rule files.
func (a *GeminiAdapter) BuildRulesIndex(modules []*domain.ModuleMemory, agentMemoryPath string, agentRulesDir string) gateway.RulesIndex {
	return buildExplicitRulesIndex(modules, agentMemoryPath, agentRulesDir)
}
