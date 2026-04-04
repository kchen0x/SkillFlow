package app

import (
	repoport "github.com/shinerio/skillflow/core/memorycatalog/app/port/repository"
	"github.com/shinerio/skillflow/core/memorycatalog/domain"
)

// MemoryService handles CRUD for memories and local push configuration.
type MemoryService struct {
	storage repoport.MemoryStorage
}

func NewMemoryService(storage repoport.MemoryStorage) *MemoryService {
	return &MemoryService{storage: storage}
}

// GetMainMemory returns the current main memory content.
func (s *MemoryService) GetMainMemory() (*domain.MainMemory, error) {
	return s.storage.GetMainMemory()
}

// SaveMainMemory saves main memory content and returns updated entity.
func (s *MemoryService) SaveMainMemory(content string) (*domain.MainMemory, error) {
	return s.storage.SaveMainMemory(content)
}

// ListModules returns all module memories sorted by name.
func (s *MemoryService) ListModules() ([]*domain.ModuleMemory, error) {
	modules, err := s.storage.ListModules()
	if err != nil {
		return nil, err
	}
	enabledMap, err := s.storage.GetAllModuleEnabled()
	if err != nil {
		return nil, err
	}
	for _, module := range modules {
		enabled, ok := enabledMap[module.Name]
		if !ok {
			enabled = true
		}
		module.Enabled = enabled
	}
	return modules, nil
}

// GetModule returns a module by name.
func (s *MemoryService) GetModule(name string) (*domain.ModuleMemory, error) {
	module, err := s.storage.GetModule(name)
	if err != nil {
		return nil, err
	}
	enabled, err := s.storage.GetModuleEnabled(name)
	if err != nil {
		return nil, err
	}
	module.Enabled = enabled == nil || *enabled
	return module, nil
}

// CreateModule creates a new module memory.
// Returns ErrModuleExists if name already taken.
func (s *MemoryService) CreateModule(name, content string) (*domain.ModuleMemory, error) {
	module, err := s.storage.CreateModule(name, content)
	if err != nil {
		return nil, err
	}
	if err := s.storage.SaveModuleEnabled(name, true); err != nil {
		return nil, err
	}
	module.Enabled = true
	return module, nil
}

// SaveModule saves module content. Returns ErrModuleNotFound if not found.
func (s *MemoryService) SaveModule(name, content string) (*domain.ModuleMemory, error) {
	module, err := s.storage.SaveModule(name, content)
	if err != nil {
		return nil, err
	}
	enabled, err := s.storage.GetModuleEnabled(name)
	if err != nil {
		return nil, err
	}
	module.Enabled = enabled == nil || *enabled
	return module, nil
}

// SetModuleEnabled updates a module's global enabled state.
func (s *MemoryService) SetModuleEnabled(name string, enabled bool) (*domain.ModuleMemory, error) {
	module, err := s.storage.GetModule(name)
	if err != nil {
		return nil, err
	}
	if err := s.storage.SaveModuleEnabled(name, enabled); err != nil {
		return nil, err
	}
	module.Enabled = enabled
	return module, nil
}

// DeleteModule deletes a module memory and removes its entry from local config.
func (s *MemoryService) DeleteModule(name string) error {
	if err := s.storage.DeleteModule(name); err != nil {
		return err
	}
	// Remove push targets entry for this module.
	if err := s.storage.DeleteModulePushTargets(name); err != nil {
		return err
	}
	if err := s.storage.DeleteModuleEnabled(name); err != nil {
		return err
	}
	return nil
}

// GetPushConfig returns push configuration for an agent (defaults if not set).
func (s *MemoryService) GetPushConfig(agentType string) (domain.MemoryPushConfig, error) {
	cfg, err := s.storage.GetPushConfig(agentType)
	if err != nil {
		return domain.MemoryPushConfig{}, err
	}
	// Apply defaults when no config is stored.
	if cfg.AgentType == "" {
		return domain.MemoryPushConfig{
			AgentType: agentType,
			Mode:      domain.PushModeMerge,
			AutoPush:  false,
		}, nil
	}
	return cfg, nil
}

// SavePushConfig saves push configuration for an agent.
func (s *MemoryService) SavePushConfig(cfg domain.MemoryPushConfig) error {
	return s.storage.SavePushConfig(cfg)
}

// GetAllPushConfigs returns push configurations for all agents.
func (s *MemoryService) GetAllPushConfigs() ([]domain.MemoryPushConfig, error) {
	return s.storage.GetAllPushConfigs()
}

// GetModulePushTargets returns push targets for a module.
func (s *MemoryService) GetModulePushTargets(moduleName string) (domain.ModulePushTargets, error) {
	targets, err := s.storage.GetModulePushTargets(moduleName)
	if err != nil {
		return domain.ModulePushTargets{}, err
	}
	// Return empty targets when none stored.
	if targets.ModuleName == "" {
		return domain.ModulePushTargets{
			ModuleName:  moduleName,
			PushTargets: []string{},
		}, nil
	}
	return targets, nil
}

// SaveModulePushTargets saves push targets for a module.
func (s *MemoryService) SaveModulePushTargets(targets domain.ModulePushTargets) error {
	return s.storage.SaveModulePushTargets(targets)
}

// GetAllModulePushTargets returns push targets for all modules.
func (s *MemoryService) GetAllModulePushTargets() ([]domain.ModulePushTargets, error) {
	return s.storage.GetAllModulePushTargets()
}
