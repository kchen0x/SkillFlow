package repository

import (
	"github.com/shinerio/skillflow/core/memorycatalog/domain"
)

// MemoryStorage provides persistence for memory content and local config.
type MemoryStorage interface {
	// Main memory
	GetMainMemory() (*domain.MainMemory, error)
	SaveMainMemory(content string) (*domain.MainMemory, error)

	// Module memories
	ListModules() ([]*domain.ModuleMemory, error)
	GetModule(name string) (*domain.ModuleMemory, error)
	CreateModule(name, content string) (*domain.ModuleMemory, error)
	SaveModule(name, content string) (*domain.ModuleMemory, error)
	DeleteModule(name string) error

	// Local config (memory_local.json)
	GetPushConfig(agentType string) (domain.MemoryPushConfig, error)
	SavePushConfig(cfg domain.MemoryPushConfig) error
	GetAllPushConfigs() ([]domain.MemoryPushConfig, error)

	GetModulePushTargets(moduleName string) (domain.ModulePushTargets, error)
	SaveModulePushTargets(targets domain.ModulePushTargets) error
	GetAllModulePushTargets() ([]domain.ModulePushTargets, error)
	DeleteModulePushTargets(moduleName string) error

	GetPushState(agentType string) (domain.MemoryPushState, error)
	SavePushState(agentType string, state domain.MemoryPushState) error
	GetAllPushStates() (map[string]domain.MemoryPushState, error)
}
