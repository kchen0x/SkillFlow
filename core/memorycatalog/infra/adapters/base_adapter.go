package adapters

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	gatewayport "github.com/shinerio/skillflow/core/memorycatalog/app/port/gateway"
	"github.com/shinerio/skillflow/core/memorycatalog/domain"
)

const (
	managedTagStart = "<skillflow-managed>"
	managedTagEnd   = "</skillflow-managed>"
	moduleTagStart  = "<skillflow-module>"
	moduleTagEnd    = "</skillflow-module>"
	sfPrefix        = "sf-"
)

type baseAdapter struct{}

// PushMainMemory writes content to agentMemoryPath.
// merge mode: manages the SkillFlow tag block inside the file
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

	newBlock := content
	startIdx, endIdx := findManagedRange(fileContent)

	var result string
	if startIdx >= 0 && endIdx >= startIdx {
		// Existing SkillFlow block found: replace it.
		result = fileContent[:startIdx] + newBlock + fileContent[endIdx:]
	} else {
		// No SkillFlow block found: append.
		if fileContent == "" {
			result = newBlock
		} else {
			result = fileContent + "\n\n" + newBlock
		}
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

// ListManagedModuleNames returns SkillFlow-managed module names from sf-*.md files.
func (b *baseAdapter) ListManagedModuleNames(agentRulesDir string) ([]string, error) {
	entries, err := os.ReadDir(agentRulesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, sfPrefix) || !strings.HasSuffix(name, ".md") {
			continue
		}
		names = append(names, strings.TrimSuffix(strings.TrimPrefix(name, sfPrefix), ".md"))
	}
	sort.Strings(names)
	return names, nil
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

// RepairManagedBlock ensures the SkillFlow tag block is intact in agentMemoryPath.
// If a managed block and optional module block are both complete: OK.
// If tags are incomplete: remove tag lines and let the next push rebuild them.
func (b *baseAdapter) RepairManagedBlock(agentMemoryPath string) error {
	data, err := os.ReadFile(agentMemoryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	fileContent := string(data)
	startIdx, endIdx := findManagedRange(fileContent)
	if startIdx >= 0 && endIdx >= startIdx {
		return nil
	}

	lines := strings.Split(fileContent, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		if isSkillFlowTagLine(line) {
			continue
		}
		filtered = append(filtered, line)
	}
	result := strings.Join(filtered, "\n")
	return os.WriteFile(agentMemoryPath, []byte(result), 0o644)
}

func isSkillFlowTagLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	return trimmed == managedTagStart ||
		trimmed == managedTagEnd ||
		trimmed == moduleTagStart ||
		trimmed == moduleTagEnd
}

func findManagedRange(fileContent string) (int, int) {
	startIdx := strings.Index(fileContent, managedTagStart)
	if startIdx < 0 {
		return -1, -1
	}

	managedEndIdx := strings.Index(fileContent[startIdx:], managedTagEnd)
	if managedEndIdx < 0 {
		return -1, -1
	}
	managedEnd := startIdx + managedEndIdx + len(managedTagEnd)

	moduleStartIdx := strings.Index(fileContent[managedEnd:], moduleTagStart)
	if moduleStartIdx < 0 {
		return startIdx, managedEnd
	}
	moduleStart := managedEnd + moduleStartIdx

	moduleEndIdx := strings.Index(fileContent[moduleStart:], moduleTagEnd)
	if moduleEndIdx < 0 {
		return -1, -1
	}
	endIdx := moduleStart + moduleEndIdx + len(moduleTagEnd)
	return startIdx, endIdx

}

// SyncConfig is a no-op for agents that require no additional config file updates.
func (b *baseAdapter) SyncConfig(_ []*domain.ModuleMemory, _ string) error {
	return nil
}

// buildExplicitRulesIndex returns a RulesIndex with explicit file listings for agents
// that do not auto-discover rules files.
func buildExplicitRulesIndex(modules []*domain.ModuleMemory, agentRulesDir string) gatewayport.RulesIndex {
	if len(modules) == 0 {
		return gatewayport.RulesIndex{}
	}
	entries := make([]string, 0, len(modules))
	for _, m := range modules {
		targetPath := filepath.Join(agentRulesDir, sfPrefix+m.Name+".md")
		entries = append(entries, "- @"+filepath.ToSlash(targetPath))
	}
	return gatewayport.RulesIndex{
		Entries: entries,
	}
}
