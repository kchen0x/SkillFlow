package adapters

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/shinerio/skillflow/core/memorycatalog/domain"
	gatewayport "github.com/shinerio/skillflow/core/memorycatalog/app/port/gateway"
)

const (
	markerStart = "<!-- SkillFlow Managed Start - DO NOT EDIT THIS BLOCK -->"
	markerEnd   = "<!-- SkillFlow Managed End -->"
	sfPrefix    = "sf-"
)

type baseAdapter struct{}

// PushMainMemory writes content to agentMemoryPath.
// merge mode: manages the marker block inside the file
// takeover mode: overwrites the entire file
func (b *baseAdapter) PushMainMemory(content string, mode domain.PushMode, agentMemoryPath string) error {
	if err := os.MkdirAll(filepath.Dir(agentMemoryPath), 0o755); err != nil {
		return err
	}

	if mode == domain.PushModeTakeover {
		return os.WriteFile(agentMemoryPath, []byte(content), 0o644)
	}

	// merge mode
	existing, err := os.ReadFile(agentMemoryPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	fileContent := string(existing)
	// Strip BOM if present
	fileContent = strings.TrimPrefix(fileContent, "\xef\xbb\xbf")

	newBlock := markerStart + "\n" + content + "\n" + markerEnd

	startIdx := strings.Index(fileContent, markerStart)
	endIdx := strings.Index(fileContent, markerEnd)

	var result string
	if startIdx >= 0 && endIdx >= 0 && endIdx > startIdx {
		// Both markers found: replace block
		result = fileContent[:startIdx] + newBlock + fileContent[endIdx+len(markerEnd):]
	} else if startIdx < 0 && endIdx < 0 {
		// Neither found: append
		if fileContent == "" {
			result = newBlock
		} else {
			result = fileContent + "\n\n" + newBlock
		}
	} else {
		// Only one marker found: repair should have fixed this, just append
		result = fileContent + "\n\n" + newBlock
	}

	return os.WriteFile(agentMemoryPath, []byte(result), 0o644)
}

// PushModuleMemory writes module content to <agentRulesDir>/sf-<name>.md
func (b *baseAdapter) PushModuleMemory(module *domain.ModuleMemory, agentRulesDir string) error {
	if err := os.MkdirAll(agentRulesDir, 0o755); err != nil {
		return err
	}

	targetPath := filepath.Join(agentRulesDir, sfPrefix+module.Name+".md")
	return os.WriteFile(targetPath, []byte(module.Content), 0o644)
}

// RemoveModuleMemory removes <agentRulesDir>/sf-<name>.md if it exists (idempotent)
func (b *baseAdapter) RemoveModuleMemory(moduleName string, agentRulesDir string) error {
	targetPath := filepath.Join(agentRulesDir, sfPrefix+moduleName+".md")
	err := os.Remove(targetPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// RepairManagedBlock ensures the marker block is intact in agentMemoryPath.
// If both markers found: OK (return nil).
// If only one found: remove that line and write file back.
// If neither found: OK (return nil).
func (b *baseAdapter) RepairManagedBlock(agentMemoryPath string) error {
	data, err := os.ReadFile(agentMemoryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	fileContent := string(data)
	startIdx := strings.Index(fileContent, markerStart)
	endIdx := strings.Index(fileContent, markerEnd)

	if startIdx >= 0 && endIdx >= 0 {
		// Both present: healthy
		return nil
	}
	if startIdx < 0 && endIdx < 0 {
		// Neither present: nothing to repair
		return nil
	}

	// Only one present: remove that line
	lines := strings.Split(fileContent, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.Contains(line, markerStart) || strings.Contains(line, markerEnd) {
			continue
		}
		filtered = append(filtered, line)
	}
	result := strings.Join(filtered, "\n")
	return os.WriteFile(agentMemoryPath, []byte(result), 0o644)
}

// buildExplicitRulesIndex returns a RulesIndex with explicit file listings for agents
// that do not auto-discover rules files.
func buildExplicitRulesIndex(modules []*domain.ModuleMemory, agentRulesDir string) gatewayport.RulesIndex {
	if len(modules) == 0 {
		return gatewayport.RulesIndex{}
	}
	entries := make([]string, 0, len(modules))
	for _, m := range modules {
		entries = append(entries, filepath.Join(agentRulesDir, sfPrefix+m.Name+".md"))
	}
	return gatewayport.RulesIndex{
		Header:  "The following rule files are managed by SkillFlow:",
		Entries: entries,
	}
}
