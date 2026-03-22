package adapters

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/shinerio/skillflow/core/memorycatalog/app/port/gateway"
	"github.com/shinerio/skillflow/core/memorycatalog/domain"
)

const openCodeConfigFile = "opencode.json"

// OpenCodeAdapter implements AgentMemoryPusher for OpenCode.
type OpenCodeAdapter struct{ *baseAdapter }

// NewOpenCodeAdapter returns a new OpenCodeAdapter.
func NewOpenCodeAdapter() *OpenCodeAdapter {
	return &OpenCodeAdapter{&baseAdapter{}}
}

// BuildRulesIndex returns empty because OpenCode loads rules via opencode.json instructions,
// not via inline refs in the AGENTS.md <skillflow-module> block.
func (a *OpenCodeAdapter) BuildRulesIndex(_ []*domain.ModuleMemory, _ string) gateway.RulesIndex {
	return gateway.RulesIndex{}
}

// SyncConfig updates opencode.json to add or remove the "rules/*.md" glob from the
// instructions whitelist depending on whether any module memories are present.
func (a *OpenCodeAdapter) SyncConfig(modules []*domain.ModuleMemory, agentRulesDir string) error {
	cleanRulesDir := filepath.Clean(agentRulesDir)
	rulesGlob := filepath.ToSlash(filepath.Join(cleanRulesDir, "*.md"))
	configDir := filepath.Dir(cleanRulesDir)
	configPath := filepath.Join(configDir, openCodeConfigFile)

	// Read existing config or start fresh.
	cfg := make(map[string]interface{})
	data, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", configPath, err)
	}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("parse %s: %w", configPath, err)
		}
	}

	// Extract current instructions slice.
	var instructions []string
	if raw, ok := cfg["instructions"]; ok {
		if arr, ok := raw.([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					instructions = append(instructions, s)
				}
			}
		}
	}

	hasModules := len(modules) > 0
	hasGlob := containsString(instructions, rulesGlob)

	switch {
	case hasModules && !hasGlob:
		instructions = append(instructions, rulesGlob)
		cfg["instructions"] = instructions
	case !hasModules && hasGlob:
		filtered := make([]string, 0, len(instructions)-1)
		for _, s := range instructions {
			if s != rulesGlob {
				filtered = append(filtered, s)
			}
		}
		if len(filtered) == 0 {
			delete(cfg, "instructions")
		} else {
			cfg["instructions"] = filtered
		}
	default:
		return nil // no change needed
	}

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", configDir, err)
	}
	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, append(out, '\n'), 0o644)
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
