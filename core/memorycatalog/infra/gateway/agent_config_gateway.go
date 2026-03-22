package gateway

import (
	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
	gatewayport "github.com/shinerio/skillflow/core/memorycatalog/app/port/gateway"
)

// ProfilesProvider is a function that returns current agent profiles.
type ProfilesProvider func() []agentdomain.AgentProfile

type AgentConfigGateway struct {
	profiles ProfilesProvider
}

func NewAgentConfigGateway(profiles ProfilesProvider) *AgentConfigGateway {
	return &AgentConfigGateway{profiles: profiles}
}

// ListEnabledAgents returns memory configs for all enabled agents that have MemoryPath set.
func (g *AgentConfigGateway) ListEnabledAgents() ([]gatewayport.AgentMemoryConfig, error) {
	var result []gatewayport.AgentMemoryConfig
	for _, p := range g.profiles() {
		if !p.Enabled || p.MemoryPath == "" {
			continue
		}
		result = append(result, gatewayport.AgentMemoryConfig{
			AgentType:  p.Name,
			MemoryPath: p.MemoryPath,
			RulesDir:   p.RulesDir,
		})
	}
	return result, nil
}

// GetAgent returns config for a specific agent type.
func (g *AgentConfigGateway) GetAgent(agentType string) (gatewayport.AgentMemoryConfig, bool, error) {
	for _, p := range g.profiles() {
		if p.Name == agentType {
			if !p.Enabled || p.MemoryPath == "" {
				return gatewayport.AgentMemoryConfig{}, false, nil
			}
			return gatewayport.AgentMemoryConfig{
				AgentType:  p.Name,
				MemoryPath: p.MemoryPath,
				RulesDir:   p.RulesDir,
			}, true, nil
		}
	}
	return gatewayport.AgentMemoryConfig{}, false, nil
}
