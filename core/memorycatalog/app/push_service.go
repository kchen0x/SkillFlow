package app

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"time"

	gatewayport "github.com/shinerio/skillflow/core/memorycatalog/app/port/gateway"
	repoport "github.com/shinerio/skillflow/core/memorycatalog/app/port/repository"
	"github.com/shinerio/skillflow/core/memorycatalog/domain"
)

// PusherResolver returns the AgentMemoryPusher for a given agentType.
type PusherResolver func(agentType string) (gatewayport.AgentMemoryPusher, bool)

// PushResult holds the outcome of a push attempt for a single agent.
type PushResult struct {
	AgentType string
	Success   bool
	Error     error
}

// PushService handles push execution to agent memory files.
type PushService struct {
	storage       repoport.MemoryStorage
	agentConfig   gatewayport.AgentConfigGateway
	resolvePusher PusherResolver
}

func NewPushService(
	storage repoport.MemoryStorage,
	agentConfig gatewayport.AgentConfigGateway,
	resolvePusher PusherResolver,
) *PushService {
	return &PushService{
		storage:       storage,
		agentConfig:   agentConfig,
		resolvePusher: resolvePusher,
	}
}

// PushToAgent pushes memory content to a single agent.
func (s *PushService) PushToAgent(agentType string) error {
	pushCfg, err := s.storage.GetPushConfig(agentType)
	if err != nil {
		return fmt.Errorf("get push config for %q: %w", agentType, err)
	}
	// Apply defaults when no config is stored.
	if pushCfg.AgentType == "" {
		pushCfg = domain.MemoryPushConfig{
			AgentType: agentType,
			Mode:      domain.PushModeMerge,
			AutoPush:  false,
		}
	}
	return s.pushModulesToAgent(agentType, nil, pushCfg.Mode)
}

// PushSelectionToAgent pushes main memory plus only the selected modules to a single agent.
func (s *PushService) PushSelectionToAgent(agentType string, moduleNames []string, mode domain.PushMode) error {
	if len(moduleNames) == 0 {
		return fmt.Errorf("no module memories selected for agent %q", agentType)
	}
	return s.pushModulesToAgent(agentType, moduleNames, mode)
}

func (s *PushService) pushModulesToAgent(agentType string, moduleNames []string, mode domain.PushMode) error {
	agentCfg, ok, err := s.agentConfig.GetAgent(agentType)
	if err != nil {
		return fmt.Errorf("get agent config for %q: %w", agentType, err)
	}
	if !ok {
		return fmt.Errorf("agent %q not found or not enabled", agentType)
	}

	mainMemory, err := s.storage.GetMainMemory()
	if err != nil {
		return fmt.Errorf("get main memory: %w", err)
	}

	allModules, err := s.storage.ListModules()
	if err != nil {
		return fmt.Errorf("list modules: %w", err)
	}

	selectedModules, removedNames, err := selectModulesForPush(allModules, moduleNames)
	if err != nil {
		return err
	}

	pusher, ok := s.resolvePusher(agentType)
	if !ok {
		return fmt.Errorf("no pusher registered for agent type %q", agentType)
	}

	if mode == domain.PushModeMerge {
		if err := pusher.RepairManagedBlock(agentCfg.MemoryPath); err != nil {
			return fmt.Errorf("repair managed block for %q: %w", agentType, err)
		}
	}

	for _, m := range selectedModules {
		if err := pusher.PushModuleMemory(m, agentCfg.RulesDir); err != nil {
			return fmt.Errorf("push module %q to agent %q: %w", m.Name, agentType, err)
		}
	}

	for _, name := range removedNames {
		if err := pusher.RemoveModuleMemory(name, agentCfg.RulesDir); err != nil {
			return fmt.Errorf("remove module %q from agent %q: %w", name, agentType, err)
		}
	}

	rulesIndex := pusher.BuildRulesIndex(selectedModules, agentCfg.RulesDir)

	composedContent := mainMemory.Content
	if len(rulesIndex.Entries) > 0 {
		var sb strings.Builder
		sb.WriteString(composedContent)
		sb.WriteString("\n\n")
		sb.WriteString(rulesIndex.Header)
		sb.WriteString("\n")
		for _, entry := range rulesIndex.Entries {
			sb.WriteString(entry)
			sb.WriteString("\n")
		}
		composedContent = sb.String()
	}

	if err := pusher.PushMainMemory(composedContent, mode, agentCfg.MemoryPath); err != nil {
		return fmt.Errorf("push main memory to agent %q: %w", agentType, err)
	}

	hash, err := computeMemoryHash(mainMemory.Content, selectedModules)
	if err != nil {
		return fmt.Errorf("compute pushed hash for agent %q: %w", agentType, err)
	}
	state := domain.MemoryPushState{
		LastPushedAt:   time.Now(),
		LastPushedHash: hash,
	}
	if err := s.storage.SavePushState(agentType, state); err != nil {
		return fmt.Errorf("save push state for agent %q: %w", agentType, err)
	}

	return nil
}

func selectModulesForPush(allModules []*domain.ModuleMemory, selectedNames []string) ([]*domain.ModuleMemory, []string, error) {
	if len(selectedNames) == 0 {
		sort.Slice(allModules, func(i, j int) bool {
			return allModules[i].Name < allModules[j].Name
		})
		return allModules, nil, nil
	}

	selectedSet := make(map[string]struct{}, len(selectedNames))
	for _, name := range selectedNames {
		selectedSet[name] = struct{}{}
	}

	var selected []*domain.ModuleMemory
	var removed []string
	for _, module := range allModules {
		if _, ok := selectedSet[module.Name]; ok {
			selected = append(selected, module)
			delete(selectedSet, module.Name)
			continue
		}
		removed = append(removed, module.Name)
	}

	if len(selectedSet) > 0 {
		missing := make([]string, 0, len(selectedSet))
		for name := range selectedSet {
			missing = append(missing, name)
		}
		sort.Strings(missing)
		return nil, nil, fmt.Errorf("selected module memories not found: %s", strings.Join(missing, ", "))
	}

	sort.Slice(selected, func(i, j int) bool {
		return selected[i].Name < selected[j].Name
	})
	sort.Strings(removed)
	return selected, removed, nil
}

// PushAll pushes memory content to all enabled agents and collects results.
func (s *PushService) PushAll() ([]PushResult, error) {
	agents, err := s.agentConfig.ListEnabledAgents()
	if err != nil {
		return nil, fmt.Errorf("list enabled agents: %w", err)
	}

	results := make([]PushResult, 0, len(agents))
	for _, agent := range agents {
		pushErr := s.PushToAgent(agent.AgentType)
		results = append(results, PushResult{
			AgentType: agent.AgentType,
			Success:   pushErr == nil,
			Error:     pushErr,
		})
	}
	return results, nil
}

// PushSelection pushes main memory plus the selected modules to multiple agents using one mode.
func (s *PushService) PushSelection(agentTypes []string, moduleNames []string, mode domain.PushMode) ([]PushResult, error) {
	results := make([]PushResult, 0, len(agentTypes))
	for _, agentType := range agentTypes {
		pushErr := s.PushSelectionToAgent(agentType, moduleNames, mode)
		results = append(results, PushResult{
			AgentType: agentType,
			Success:   pushErr == nil,
			Error:     pushErr,
		})
	}
	return results, nil
}

// ComputeAgentHash computes a SHA256 hash of the main memory content plus
// all module contents, sorted by module name.
func (s *PushService) ComputeAgentHash(agentType string) (string, error) {
	mainMemory, err := s.storage.GetMainMemory()
	if err != nil {
		return "", fmt.Errorf("get main memory: %w", err)
	}

	allModules, err := s.storage.ListModules()
	if err != nil {
		return "", fmt.Errorf("list modules: %w", err)
	}
	return computeMemoryHash(mainMemory.Content, allModules)
}

func computeMemoryHash(mainContent string, modules []*domain.ModuleMemory) (string, error) {
	sortedModules := append([]*domain.ModuleMemory(nil), modules...)
	sort.Slice(sortedModules, func(i, j int) bool {
		return sortedModules[i].Name < sortedModules[j].Name
	})

	h := sha256.New()
	if _, err := h.Write([]byte(mainContent)); err != nil {
		return "", err
	}
	for _, module := range sortedModules {
		if _, err := h.Write([]byte("\n")); err != nil {
			return "", err
		}
		if _, err := h.Write([]byte(module.Content)); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// GetPushStatus returns one of: "synced", "pendingPush", "neverPushed".
func (s *PushService) GetPushStatus(agentType string) (string, error) {
	state, err := s.storage.GetPushState(agentType)
	if err != nil {
		return "", fmt.Errorf("get push state for %q: %w", agentType, err)
	}
	if state.LastPushedAt.IsZero() {
		return "neverPushed", nil
	}
	currentHash, err := s.ComputeAgentHash(agentType)
	if err != nil {
		return "", fmt.Errorf("compute hash for %q: %w", agentType, err)
	}
	if currentHash == state.LastPushedHash {
		return "synced", nil
	}
	return "pendingPush", nil
}

// GetAllPushStatuses returns push status for all enabled agents.
func (s *PushService) GetAllPushStatuses() (map[string]string, error) {
	agents, err := s.agentConfig.ListEnabledAgents()
	if err != nil {
		return nil, fmt.Errorf("list enabled agents: %w", err)
	}

	statuses := make(map[string]string, len(agents))
	for _, agent := range agents {
		status, err := s.GetPushStatus(agent.AgentType)
		if err != nil {
			return nil, err
		}
		statuses[agent.AgentType] = status
	}
	return statuses, nil
}
