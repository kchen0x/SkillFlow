package main

import (
	"fmt"
	"path/filepath"

	memorydomain "github.com/shinerio/skillflow/core/memorycatalog/domain"
	memoryeditor "github.com/shinerio/skillflow/core/memorycatalog/infra/editor"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ── DTOs ────────────────────────────────────────────────────────────────────

// MainMemoryDTO is returned to the frontend for main memory.
type MainMemoryDTO struct {
	Content   string `json:"content"`
	UpdatedAt string `json:"updatedAt"` // RFC3339
}

// ModuleMemoryDTO is returned to the frontend for a module memory.
type ModuleMemoryDTO struct {
	Name      string `json:"name"`
	Content   string `json:"content"`
	UpdatedAt string `json:"updatedAt"` // RFC3339
}

// MemoryPushConfigDTO is returned to the frontend for per-agent push config.
type MemoryPushConfigDTO struct {
	AgentType string `json:"agentType"`
	Mode      string `json:"mode"` // "merge" or "takeover"
	AutoPush  bool   `json:"autoPush"`
}

// ModulePushTargetsDTO is returned to the frontend for per-module push targets.
type ModulePushTargetsDTO struct {
	ModuleName  string   `json:"moduleName"`
	PushTargets []string `json:"pushTargets"`
}

// PushStatusDTO is returned to the frontend for per-agent push status.
type PushStatusDTO struct {
	AgentType string `json:"agentType"`
	Status    string `json:"status"` // "synced" | "pendingPush" | "neverPushed"
}

// PushResultDTO is returned to the frontend for a single push result.
type PushResultDTO struct {
	AgentType string `json:"agentType"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

// ── Main memory ──────────────────────────────────────────────────────────────

// GetMainMemory returns the current main memory content.
func (a *App) GetMainMemory() (*MainMemoryDTO, error) {
	m, err := a.memoryService.GetMainMemory()
	if err != nil {
		return nil, err
	}
	return &MainMemoryDTO{
		Content:   m.Content,
		UpdatedAt: m.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// SaveMainMemory saves main memory content and returns the updated entity.
func (a *App) SaveMainMemory(content string) (*MainMemoryDTO, error) {
	m, err := a.memoryService.SaveMainMemory(content)
	if err != nil {
		return nil, err
	}
	a.emitMemoryEvent(EventMemoryContentChanged, map[string]interface{}{
		"type": "main",
	})
	if err := a.syncMemoryToAutoPushAgents(); err != nil {
		a.logErrorf("memory auto sync failed after save main memory: %v", err)
	}
	return &MainMemoryDTO{
		Content:   m.Content,
		UpdatedAt: m.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// ── Module memories ──────────────────────────────────────────────────────────

// ListModuleMemories returns all module memories sorted by name.
func (a *App) ListModuleMemories() ([]*ModuleMemoryDTO, error) {
	modules, err := a.memoryService.ListModules()
	if err != nil {
		return nil, err
	}
	dtos := make([]*ModuleMemoryDTO, 0, len(modules))
	for _, m := range modules {
		dtos = append(dtos, &ModuleMemoryDTO{
			Name:      m.Name,
			Content:   m.Content,
			UpdatedAt: m.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return dtos, nil
}

// GetModuleMemory returns a single module memory by name.
func (a *App) GetModuleMemory(name string) (*ModuleMemoryDTO, error) {
	m, err := a.memoryService.GetModule(name)
	if err != nil {
		return nil, err
	}
	return &ModuleMemoryDTO{
		Name:      m.Name,
		Content:   m.Content,
		UpdatedAt: m.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// CreateModuleMemory creates a new module memory.
func (a *App) CreateModuleMemory(name, content string) (*ModuleMemoryDTO, error) {
	m, err := a.memoryService.CreateModule(name, content)
	if err != nil {
		return nil, err
	}
	a.emitMemoryEvent(EventMemoryContentChanged, map[string]interface{}{
		"type": "module",
		"name": name,
	})
	if err := a.syncMemoryToAutoPushAgents(); err != nil {
		a.logErrorf("memory auto sync failed after create module: module=%s err=%v", name, err)
	}
	return &ModuleMemoryDTO{
		Name:      m.Name,
		Content:   m.Content,
		UpdatedAt: m.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// SaveModuleMemory saves module content.
func (a *App) SaveModuleMemory(name, content string) (*ModuleMemoryDTO, error) {
	m, err := a.memoryService.SaveModule(name, content)
	if err != nil {
		return nil, err
	}
	a.emitMemoryEvent(EventMemoryContentChanged, map[string]interface{}{
		"type": "module",
		"name": name,
	})
	if err := a.syncMemoryToAutoPushAgents(); err != nil {
		a.logErrorf("memory auto sync failed after save module: module=%s err=%v", name, err)
	}
	return &ModuleMemoryDTO{
		Name:      m.Name,
		Content:   m.Content,
		UpdatedAt: m.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// DeleteModuleMemory deletes a module memory and auto-syncs enabled agents.
func (a *App) DeleteModuleMemory(name string) error {
	if err := a.memoryService.DeleteModule(name); err != nil {
		return err
	}
	a.emitMemoryEvent(EventMemoryContentChanged, map[string]interface{}{
		"type": "module",
		"name": name,
	})
	if err := a.syncMemoryToAutoPushAgents(); err != nil {
		a.logErrorf("memory auto sync failed after delete module: module=%s err=%v", name, err)
	}
	return nil
}

// ── Push configuration ────────────────────────────────────────────────────────

// GetMemoryPushConfig returns push configuration for an agent (defaults if not set).
func (a *App) GetMemoryPushConfig(agentType string) (*MemoryPushConfigDTO, error) {
	cfg, err := a.memoryService.GetPushConfig(agentType)
	if err != nil {
		return nil, err
	}
	return &MemoryPushConfigDTO{
		AgentType: cfg.AgentType,
		Mode:      string(cfg.Mode),
		AutoPush:  cfg.AutoPush,
	}, nil
}

// SaveMemoryPushConfig saves push configuration for an agent.
func (a *App) SaveMemoryPushConfig(agentType, mode string, autoPush bool) error {
	cfg := memorydomain.MemoryPushConfig{
		AgentType: agentType,
		Mode:      memorydomain.PushMode(mode),
		AutoPush:  autoPush,
	}
	return a.memoryService.SavePushConfig(cfg)
}

// GetAllMemoryPushConfigs returns push configurations for all agents that have been configured.
func (a *App) GetAllMemoryPushConfigs() ([]*MemoryPushConfigDTO, error) {
	cfgs, err := a.memoryService.GetAllPushConfigs()
	if err != nil {
		return nil, err
	}
	dtos := make([]*MemoryPushConfigDTO, 0, len(cfgs))
	for _, c := range cfgs {
		dtos = append(dtos, &MemoryPushConfigDTO{
			AgentType: c.AgentType,
			Mode:      string(c.Mode),
			AutoPush:  c.AutoPush,
		})
	}
	return dtos, nil
}

// ── Module push targets ───────────────────────────────────────────────────────

// GetModulePushTargets returns push targets for a module.
func (a *App) GetModulePushTargets(moduleName string) (*ModulePushTargetsDTO, error) {
	t, err := a.memoryService.GetModulePushTargets(moduleName)
	if err != nil {
		return nil, err
	}
	return &ModulePushTargetsDTO{
		ModuleName:  t.ModuleName,
		PushTargets: t.PushTargets,
	}, nil
}

// SaveModulePushTargets saves push targets for a module.
func (a *App) SaveModulePushTargets(moduleName string, pushTargets []string) error {
	return a.memoryService.SaveModulePushTargets(memorydomain.ModulePushTargets{
		ModuleName:  moduleName,
		PushTargets: pushTargets,
	})
}

// GetAllModulePushTargets returns push targets for all modules.
func (a *App) GetAllModulePushTargets() ([]*ModulePushTargetsDTO, error) {
	all, err := a.memoryService.GetAllModulePushTargets()
	if err != nil {
		return nil, err
	}
	dtos := make([]*ModulePushTargetsDTO, 0, len(all))
	for _, t := range all {
		dtos = append(dtos, &ModulePushTargetsDTO{
			ModuleName:  t.ModuleName,
			PushTargets: t.PushTargets,
		})
	}
	return dtos, nil
}

// ── Push operations ───────────────────────────────────────────────────────────

// PushMemoryToAgent pushes memory content to a single agent.
func (a *App) PushMemoryToAgent(agentType string) (*PushResultDTO, error) {
	err := a.memoryPushService.PushToAgent(agentType)
	success := err == nil
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	a.emitMemoryEvent(EventMemoryPushCompleted, map[string]interface{}{
		"agent":   agentType,
		"success": success,
	})
	status := a.memoryStatusForAgent(agentType)
	a.emitMemoryEvent(EventMemoryStatusChanged, map[string]interface{}{
		"agent":  agentType,
		"status": status,
	})
	return &PushResultDTO{
		AgentType: agentType,
		Success:   success,
		Error:     errStr,
	}, nil
}

// PushAllMemory pushes memory content to all enabled agents.
func (a *App) PushAllMemory() ([]*PushResultDTO, error) {
	results, err := a.memoryPushService.PushAll()
	if err != nil {
		return nil, err
	}
	dtos := make([]*PushResultDTO, 0, len(results))
	for _, r := range results {
		errStr := ""
		if r.Error != nil {
			errStr = r.Error.Error()
		}
		a.emitMemoryEvent(EventMemoryPushCompleted, map[string]interface{}{
			"agent":   r.AgentType,
			"success": r.Success,
		})
		status := a.memoryStatusForAgent(r.AgentType)
		a.emitMemoryEvent(EventMemoryStatusChanged, map[string]interface{}{
			"agent":  r.AgentType,
			"status": status,
		})
		dtos = append(dtos, &PushResultDTO{
			AgentType: r.AgentType,
			Success:   r.Success,
			Error:     errStr,
		})
	}
	return dtos, nil
}

// PushSelectedMemory pushes main memory plus the selected modules to target agents.
func (a *App) PushSelectedMemory(agentTypes []string, moduleNames []string, mode string) ([]*PushResultDTO, error) {
	results, err := a.memoryPushService.PushSelection(agentTypes, moduleNames, memorydomain.PushMode(mode))
	if err != nil {
		return nil, err
	}

	dtos := make([]*PushResultDTO, 0, len(results))
	for _, r := range results {
		errStr := ""
		if r.Error != nil {
			errStr = r.Error.Error()
		}
		a.emitMemoryEvent(EventMemoryPushCompleted, map[string]interface{}{
			"agent":   r.AgentType,
			"success": r.Success,
		})
		a.emitMemoryEvent(EventMemoryStatusChanged, map[string]interface{}{
			"agent":  r.AgentType,
			"status": a.memoryStatusForAgent(r.AgentType),
		})
		dtos = append(dtos, &PushResultDTO{
			AgentType: r.AgentType,
			Success:   r.Success,
			Error:     errStr,
		})
	}
	return dtos, nil
}

// ── Push status ───────────────────────────────────────────────────────────────

// GetMemoryPushStatus returns push status for a single agent.
func (a *App) GetMemoryPushStatus(agentType string) (*PushStatusDTO, error) {
	status, err := a.memoryPushService.GetPushStatus(agentType)
	if err != nil {
		return nil, err
	}
	return &PushStatusDTO{
		AgentType: agentType,
		Status:    status,
	}, nil
}

// GetAllMemoryPushStatuses returns push status for all enabled agents.
func (a *App) GetAllMemoryPushStatuses() ([]*PushStatusDTO, error) {
	statuses, err := a.memoryPushService.GetAllPushStatuses()
	if err != nil {
		return nil, err
	}
	dtos := make([]*PushStatusDTO, 0, len(statuses))
	for agentType, status := range statuses {
		dtos = append(dtos, &PushStatusDTO{
			AgentType: agentType,
			Status:    status,
		})
	}
	return dtos, nil
}

// ── Editor ────────────────────────────────────────────────────────────────────

// OpenMemoryInEditor opens a memory file in the system default editor.
// memoryType = "main" opens main.md; memoryType = "module" opens rules/<moduleName>.md.
func (a *App) OpenMemoryInEditor(memoryType string, moduleName string) error {
	var path string
	switch memoryType {
	case "main":
		path = filepath.Join(a.dataDir(), "memory", "main.md")
	case "module":
		if moduleName == "" {
			return fmt.Errorf("moduleName is required for memoryType %q", memoryType)
		}
		path = filepath.Join(a.dataDir(), "memory", "rules", moduleName+".md")
	default:
		return fmt.Errorf("unknown memoryType %q", memoryType)
	}
	if err := memoryeditor.OpenFile(path); err != nil {
		return err
	}
	a.watchExternalMemoryChanges(memoryType, moduleName, path)
	return nil
}

func (a *App) syncMemoryToAutoPushAgents() error {
	agents, err := a.GetEnabledAgents()
	if err != nil {
		return err
	}

	autoPushAgents := make([]string, 0, len(agents))
	for _, agent := range agents {
		cfg, cfgErr := a.memoryService.GetPushConfig(agent.Name)
		if cfgErr != nil {
			return cfgErr
		}
		if cfg.AutoPush {
			autoPushAgents = append(autoPushAgents, agent.Name)
		}
	}
	if len(autoPushAgents) == 0 {
		return nil
	}

	a.logInfof("memory auto sync started: agentCount=%d", len(autoPushAgents))
	var firstErr error
	for _, agentType := range autoPushAgents {
		pushErr := a.memoryPushService.PushToAgent(agentType)
		success := pushErr == nil
		if pushErr != nil {
			a.logErrorf("memory auto sync failed: agent=%s err=%v", agentType, pushErr)
			if firstErr == nil {
				firstErr = pushErr
			}
		}
		a.emitMemoryEvent(EventMemoryPushCompleted, map[string]interface{}{
			"agent":   agentType,
			"success": success,
		})
		a.emitMemoryEvent(EventMemoryStatusChanged, map[string]interface{}{
			"agent":  agentType,
			"status": a.memoryStatusForAgent(agentType),
		})
	}
	if firstErr == nil {
		a.logInfof("memory auto sync completed: agentCount=%d", len(autoPushAgents))
	}
	return firstErr
}

func (a *App) memoryStatusForAgent(agentType string) string {
	status, err := a.memoryPushService.GetPushStatus(agentType)
	if err != nil {
		a.logErrorf("memory status load failed: agent=%s err=%v", agentType, err)
		return "pendingPush"
	}
	return status
}

func (a *App) emitMemoryEvent(eventName string, payload map[string]interface{}) {
	if a == nil || a.ctx == nil {
		return
	}
	runtime.EventsEmit(a.ctx, eventName, payload)
}
