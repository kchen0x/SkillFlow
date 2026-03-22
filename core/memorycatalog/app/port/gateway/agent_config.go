package gateway

// AgentConfigGateway reads agent configuration from agentintegration context.
type AgentConfigGateway interface {
	ListEnabledAgents() ([]AgentMemoryConfig, error)
	GetAgent(agentType string) (AgentMemoryConfig, bool, error)
}

// AgentMemoryConfig is the cross-context read DTO for agent memory paths.
type AgentMemoryConfig struct {
	AgentType  string
	MemoryPath string // e.g. ~/.claude/CLAUDE.md
	RulesDir   string // e.g. ~/.claude/rules/
}
