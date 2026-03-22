package gateway

import "github.com/shinerio/skillflow/core/memorycatalog/domain"

// RulesIndex is a structured representation of the rules index block.
// Claude Code returns empty (auto-scan), others return listing.
type RulesIndex struct {
	Header  string   // e.g. "The following rule files are managed by SkillFlow:"
	Entries []string // absolute paths to sf-*.md files
}

// AgentMemoryPusher handles the filesystem details of writing to agent directories.
// Implemented per agent type in infra/adapters/.
type AgentMemoryPusher interface {
	// PushMainMemory pushes main memory content to agent's main memory file.
	PushMainMemory(content string, mode domain.PushMode, agentMemoryPath string) error
	// PushModuleMemory pushes module memory to agent's rules directory (writes sf-<name>.md).
	PushModuleMemory(module *domain.ModuleMemory, agentRulesDir string) error
	// RemoveModuleMemory removes a pushed module memory from agent's rules directory.
	RemoveModuleMemory(moduleName string, agentRulesDir string) error
	// BuildRulesIndex builds rules index block (Claude Code returns empty, others return listing).
	BuildRulesIndex(modules []*domain.ModuleMemory, agentRulesDir string) RulesIndex
	// RepairManagedBlock detects and repairs corrupted marker blocks in merge mode.
	RepairManagedBlock(agentMemoryPath string) error
}
