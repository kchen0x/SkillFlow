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
	// 1. Get agent config.
	agentCfg, ok, err := s.agentConfig.GetAgent(agentType)
	if err != nil {
		return fmt.Errorf("get agent config for %q: %w", agentType, err)
	}
	if !ok {
		return fmt.Errorf("agent %q not found or not enabled", agentType)
	}

	// 2. Get push config for agent.
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

	// 3. Get main memory content.
	mainMemory, err := s.storage.GetMainMemory()
	if err != nil {
		return fmt.Errorf("get main memory: %w", err)
	}

	// 4. Get all modules and filter to those targeting this agent.
	allModules, err := s.storage.ListModules()
	if err != nil {
		return fmt.Errorf("list modules: %w", err)
	}
	allTargets, err := s.storage.GetAllModulePushTargets()
	if err != nil {
		return fmt.Errorf("get all module push targets: %w", err)
	}

	// Build a set of module names targeting this agent.
	targetedNames := make(map[string]bool)
	for _, t := range allTargets {
		for _, at := range t.PushTargets {
			if at == agentType {
				targetedNames[t.ModuleName] = true
				break
			}
		}
	}

	// Separate modules into targeted and not-targeted.
	var targetedModules []*domain.ModuleMemory
	var untargetedNames []string
	for _, m := range allModules {
		if targetedNames[m.Name] {
			targetedModules = append(targetedModules, m)
		} else {
			untargetedNames = append(untargetedNames, m.Name)
		}
	}

	// 5. Get pusher from resolver.
	pusher, ok := s.resolvePusher(agentType)
	if !ok {
		return fmt.Errorf("no pusher registered for agent type %q", agentType)
	}

	// 6. Repair managed block first (merge mode only).
	if pushCfg.Mode == domain.PushModeMerge {
		if err := pusher.RepairManagedBlock(agentCfg.MemoryPath); err != nil {
			return fmt.Errorf("repair managed block for %q: %w", agentType, err)
		}
	}

	// 7. Push each targeted module.
	for _, m := range targetedModules {
		if err := pusher.PushModuleMemory(m, agentCfg.RulesDir); err != nil {
			return fmt.Errorf("push module %q to agent %q: %w", m.Name, agentType, err)
		}
	}

	// 8. Remove modules no longer targeting this agent.
	for _, name := range untargetedNames {
		if err := pusher.RemoveModuleMemory(name, agentCfg.RulesDir); err != nil {
			return fmt.Errorf("remove module %q from agent %q: %w", name, agentType, err)
		}
	}

	// 9. Build rules index.
	rulesIndex := pusher.BuildRulesIndex(targetedModules, agentCfg.RulesDir)

	// 10. Compose main memory push content.
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

	// 11. Push main memory.
	if err := pusher.PushMainMemory(composedContent, pushCfg.Mode, agentCfg.MemoryPath); err != nil {
		return fmt.Errorf("push main memory to agent %q: %w", agentType, err)
	}

	// 12. Compute per-agent content hash and save push state.
	hash, err := s.ComputeAgentHash(agentType)
	if err != nil {
		return fmt.Errorf("compute hash for agent %q: %w", agentType, err)
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

// ComputeAgentHash computes a SHA256 hash of the main memory content plus
// all module contents targeted at agentType, sorted by module name.
func (s *PushService) ComputeAgentHash(agentType string) (string, error) {
	mainMemory, err := s.storage.GetMainMemory()
	if err != nil {
		return "", fmt.Errorf("get main memory: %w", err)
	}

	allModules, err := s.storage.ListModules()
	if err != nil {
		return "", fmt.Errorf("list modules: %w", err)
	}
	allTargets, err := s.storage.GetAllModulePushTargets()
	if err != nil {
		return "", fmt.Errorf("get all module push targets: %w", err)
	}

	targetedNames := make(map[string]bool)
	for _, t := range allTargets {
		for _, at := range t.PushTargets {
			if at == agentType {
				targetedNames[t.ModuleName] = true
				break
			}
		}
	}

	var targetedModules []*domain.ModuleMemory
	for _, m := range allModules {
		if targetedNames[m.Name] {
			targetedModules = append(targetedModules, m)
		}
	}
	sort.Slice(targetedModules, func(i, j int) bool {
		return targetedModules[i].Name < targetedModules[j].Name
	})

	h := sha256.New()
	h.Write([]byte(mainMemory.Content))
	for _, m := range targetedModules {
		h.Write([]byte("\n"))
		h.Write([]byte(m.Content))
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
