package repository

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/shinerio/skillflow/core/memorycatalog/domain"
	"github.com/shinerio/skillflow/core/platform/appdata"
)

// FsStorage implements MemoryStorage using the filesystem.
// Layout:
//
//	<memoryDir>/main.md
//	<memoryDir>/rules/<name>.md
//	<memoryDir>/memory_local.json
type FsStorage struct {
	memoryDir string
	mu        sync.Mutex
}

// NewFsStorage creates a new FsStorage rooted at <dataDir>/memory.
func NewFsStorage(dataDir string) *FsStorage {
	return &FsStorage{memoryDir: appdata.MemoryDir(dataDir)}
}

// ── helpers ──────────────────────────────────────────────────────────────────

func (s *FsStorage) mainPath() string {
	return filepath.Join(s.memoryDir, "main.md")
}

func (s *FsStorage) rulesDir() string {
	return filepath.Join(s.memoryDir, "rules")
}

func (s *FsStorage) modulePath(name string) string {
	return filepath.Join(s.rulesDir(), name+".md")
}

func (s *FsStorage) localConfigPath() string {
	return filepath.Join(s.memoryDir, "memory_local.json")
}

// readUTF8 reads a file and strips a leading UTF-8 BOM if present.
func readUTF8(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) {
		data = data[3:]
	}
	return string(data), nil
}

// writeAtomic writes data atomically via a temp file then rename.
func writeAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".tmp_*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		tmp.Close()
		os.Remove(tmpName)
	}()
	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

// fileMtime returns the modification time of a file.
func fileMtime(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

// ── Main memory ──────────────────────────────────────────────────────────────

// GetMainMemory reads main.md. If the file does not exist it returns an empty
// MainMemory (not an error).
func (s *FsStorage) GetMainMemory() (*domain.MainMemory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	content, err := readUTF8(s.mainPath())
	if errors.Is(err, os.ErrNotExist) {
		return &domain.MainMemory{Content: "", UpdatedAt: time.Time{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get main memory: %w", err)
	}
	return &domain.MainMemory{
		Content:   content,
		UpdatedAt: fileMtime(s.mainPath()),
	}, nil
}

// SaveMainMemory writes content to main.md, creating directories as needed.
func (s *FsStorage) SaveMainMemory(content string) (*domain.MainMemory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := writeAtomic(s.mainPath(), []byte(content)); err != nil {
		return nil, fmt.Errorf("save main memory: %w", err)
	}
	return &domain.MainMemory{
		Content:   content,
		UpdatedAt: fileMtime(s.mainPath()),
	}, nil
}

// ── Module memories ──────────────────────────────────────────────────────────

// ListModules returns all module memories sorted by name. Returns an empty
// slice if the rules directory does not exist.
func (s *FsStorage) ListModules() ([]*domain.ModuleMemory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := os.ReadDir(s.rulesDir())
	if errors.Is(err, os.ErrNotExist) {
		return []*domain.ModuleMemory{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list modules: %w", err)
	}

	var modules []*domain.ModuleMemory
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".md")
		path := s.modulePath(name)
		content, err := readUTF8(path)
		if err != nil {
			return nil, fmt.Errorf("list modules: read %q: %w", name, err)
		}
		modules = append(modules, &domain.ModuleMemory{
			Name:      name,
			Content:   content,
			Enabled:   true,
			UpdatedAt: fileMtime(path),
		})
	}

	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Name < modules[j].Name
	})
	return modules, nil
}

// GetModule reads a single module by name. Returns domain.ErrModuleNotFound
// if the file does not exist.
func (s *FsStorage) GetModule(name string) (*domain.ModuleMemory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.modulePath(name)
	content, err := readUTF8(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, domain.ErrModuleNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get module %q: %w", name, err)
	}
	return &domain.ModuleMemory{
		Name:      name,
		Content:   content,
		Enabled:   true,
		UpdatedAt: fileMtime(path),
	}, nil
}

// CreateModule creates a new module file. Returns domain.ErrInvalidModuleName
// for an invalid name and domain.ErrModuleExists if the file already exists.
func (s *FsStorage) CreateModule(name, content string) (*domain.ModuleMemory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := domain.ValidateModuleName(name); err != nil {
		return nil, err
	}

	path := s.modulePath(name)
	if _, err := os.Stat(path); err == nil {
		return nil, domain.ErrModuleExists
	}

	if err := writeAtomic(path, []byte(content)); err != nil {
		return nil, fmt.Errorf("create module %q: %w", name, err)
	}
	return &domain.ModuleMemory{
		Name:      name,
		Content:   content,
		Enabled:   true,
		UpdatedAt: fileMtime(path),
	}, nil
}

// SaveModule overwrites an existing module file. Returns domain.ErrModuleNotFound
// if the file does not exist.
func (s *FsStorage) SaveModule(name, content string) (*domain.ModuleMemory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.modulePath(name)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil, domain.ErrModuleNotFound
	}

	if err := writeAtomic(path, []byte(content)); err != nil {
		return nil, fmt.Errorf("save module %q: %w", name, err)
	}
	return &domain.ModuleMemory{
		Name:      name,
		Content:   content,
		Enabled:   true,
		UpdatedAt: fileMtime(path),
	}, nil
}

// DeleteModule removes a module file. Returns domain.ErrModuleNotFound if the
// file does not exist.
func (s *FsStorage) DeleteModule(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.modulePath(name)
	err := os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return domain.ErrModuleNotFound
	}
	if err != nil {
		return fmt.Errorf("delete module %q: %w", name, err)
	}
	return nil
}

// ── Local config JSON structs ─────────────────────────────────────────────────

type localConfig struct {
	PushConfigs  map[string]pushConfigJSON    `json:"pushConfigs"`
	Modules      map[string]moduleTargetsJSON `json:"modules"`
	ModuleStates map[string]moduleStateJSON   `json:"moduleStates"`
	PushState    map[string]pushStateJSON     `json:"pushState"`
}

type pushConfigJSON struct {
	Mode     string `json:"mode"`
	AutoPush bool   `json:"autoPush"`
}

type moduleTargetsJSON struct {
	PushTargets []string `json:"pushTargets"`
}

type moduleStateJSON struct {
	Enabled bool `json:"enabled"`
}

type pushStateJSON struct {
	LastPushedAt   time.Time `json:"lastPushedAt"`
	LastPushedHash string    `json:"lastPushedHash"`
}

// ── Local config I/O ──────────────────────────────────────────────────────────

// loadLocalConfig reads memory_local.json. Caller must hold mu.
func (s *FsStorage) loadLocalConfig() (localConfig, error) {
	var cfg localConfig
	data, err := os.ReadFile(s.localConfigPath())
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, fmt.Errorf("load local config: %w", err)
	}
	if bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) {
		data = data[3:]
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse local config: %w", err)
	}
	return cfg, nil
}

// saveLocalConfig writes memory_local.json. Caller must hold mu.
func (s *FsStorage) saveLocalConfig(cfg localConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal local config: %w", err)
	}
	if err := writeAtomic(s.localConfigPath(), data); err != nil {
		return fmt.Errorf("save local config: %w", err)
	}
	return nil
}

// ── Push configs ──────────────────────────────────────────────────────────────

// GetPushConfig returns the push config for the given agent type. Returns an
// empty MemoryPushConfig (not an error) if not found — callers apply defaults.
func (s *FsStorage) GetPushConfig(agentType string) (domain.MemoryPushConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.loadLocalConfig()
	if err != nil {
		return domain.MemoryPushConfig{}, err
	}
	raw, ok := cfg.PushConfigs[agentType]
	if !ok {
		return domain.MemoryPushConfig{}, nil
	}
	return domain.MemoryPushConfig{
		AgentType: agentType,
		Mode:      domain.PushMode(raw.Mode),
		AutoPush:  raw.AutoPush,
	}, nil
}

// SavePushConfig persists the push config for the agent type in the config.
func (s *FsStorage) SavePushConfig(pc domain.MemoryPushConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.loadLocalConfig()
	if err != nil {
		return err
	}
	if cfg.PushConfigs == nil {
		cfg.PushConfigs = make(map[string]pushConfigJSON)
	}
	cfg.PushConfigs[pc.AgentType] = pushConfigJSON{
		Mode:     string(pc.Mode),
		AutoPush: pc.AutoPush,
	}
	return s.saveLocalConfig(cfg)
}

// GetAllPushConfigs returns all stored push configs as a slice.
func (s *FsStorage) GetAllPushConfigs() ([]domain.MemoryPushConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.loadLocalConfig()
	if err != nil {
		return nil, err
	}
	result := make([]domain.MemoryPushConfig, 0, len(cfg.PushConfigs))
	for agentType, raw := range cfg.PushConfigs {
		result = append(result, domain.MemoryPushConfig{
			AgentType: agentType,
			Mode:      domain.PushMode(raw.Mode),
			AutoPush:  raw.AutoPush,
		})
	}
	return result, nil
}

// ── Module push targets ────────────────────────────────────────────────────────

// GetModulePushTargets returns the push targets for the given module. Returns
// an empty ModulePushTargets (not an error) if not found.
func (s *FsStorage) GetModulePushTargets(moduleName string) (domain.ModulePushTargets, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.loadLocalConfig()
	if err != nil {
		return domain.ModulePushTargets{}, err
	}
	raw, ok := cfg.Modules[moduleName]
	if !ok {
		return domain.ModulePushTargets{}, nil
	}
	targets := make([]string, len(raw.PushTargets))
	copy(targets, raw.PushTargets)
	return domain.ModulePushTargets{
		ModuleName:  moduleName,
		PushTargets: targets,
	}, nil
}

// SaveModulePushTargets persists the module push targets in the config.
func (s *FsStorage) SaveModulePushTargets(targets domain.ModulePushTargets) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.loadLocalConfig()
	if err != nil {
		return err
	}
	if cfg.Modules == nil {
		cfg.Modules = make(map[string]moduleTargetsJSON)
	}
	pts := make([]string, len(targets.PushTargets))
	copy(pts, targets.PushTargets)
	cfg.Modules[targets.ModuleName] = moduleTargetsJSON{PushTargets: pts}
	return s.saveLocalConfig(cfg)
}

// GetAllModulePushTargets returns all stored module push targets as a slice.
func (s *FsStorage) GetAllModulePushTargets() ([]domain.ModulePushTargets, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.loadLocalConfig()
	if err != nil {
		return nil, err
	}
	result := make([]domain.ModulePushTargets, 0, len(cfg.Modules))
	for moduleName, raw := range cfg.Modules {
		pts := make([]string, len(raw.PushTargets))
		copy(pts, raw.PushTargets)
		result = append(result, domain.ModulePushTargets{
			ModuleName:  moduleName,
			PushTargets: pts,
		})
	}
	return result, nil
}

// DeleteModulePushTargets removes the push targets entry for a module.
func (s *FsStorage) DeleteModulePushTargets(moduleName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.loadLocalConfig()
	if err != nil {
		return err
	}
	if cfg.Modules == nil {
		return nil
	}
	delete(cfg.Modules, moduleName)
	return s.saveLocalConfig(cfg)
}

// GetModuleEnabled returns the stored enabled state for a module. Nil means unset.
func (s *FsStorage) GetModuleEnabled(moduleName string) (*bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.loadLocalConfig()
	if err != nil {
		return nil, err
	}
	if cfg.ModuleStates == nil {
		return nil, nil
	}
	raw, ok := cfg.ModuleStates[moduleName]
	if !ok {
		return nil, nil
	}
	enabled := raw.Enabled
	return &enabled, nil
}

// SaveModuleEnabled persists the enabled state for a module.
func (s *FsStorage) SaveModuleEnabled(moduleName string, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.loadLocalConfig()
	if err != nil {
		return err
	}
	if cfg.ModuleStates == nil {
		cfg.ModuleStates = make(map[string]moduleStateJSON)
	}
	cfg.ModuleStates[moduleName] = moduleStateJSON{Enabled: enabled}
	return s.saveLocalConfig(cfg)
}

// GetAllModuleEnabled returns all stored module enabled states.
func (s *FsStorage) GetAllModuleEnabled() (map[string]bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.loadLocalConfig()
	if err != nil {
		return nil, err
	}
	result := make(map[string]bool, len(cfg.ModuleStates))
	for moduleName, raw := range cfg.ModuleStates {
		result[moduleName] = raw.Enabled
	}
	return result, nil
}

// DeleteModuleEnabled removes the enabled state entry for a module.
func (s *FsStorage) DeleteModuleEnabled(moduleName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.loadLocalConfig()
	if err != nil {
		return err
	}
	if cfg.ModuleStates == nil {
		return nil
	}
	delete(cfg.ModuleStates, moduleName)
	return s.saveLocalConfig(cfg)
}

// ── Push state ─────────────────────────────────────────────────────────────────

// GetPushState returns the push state for the given agent type. Returns an
// empty MemoryPushState (not an error) if not found.
func (s *FsStorage) GetPushState(agentType string) (domain.MemoryPushState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.loadLocalConfig()
	if err != nil {
		return domain.MemoryPushState{}, err
	}
	raw, ok := cfg.PushState[agentType]
	if !ok {
		return domain.MemoryPushState{}, nil
	}
	return domain.MemoryPushState{
		LastPushedAt:   raw.LastPushedAt,
		LastPushedHash: raw.LastPushedHash,
	}, nil
}

// SavePushState persists the push state for the given agent type.
func (s *FsStorage) SavePushState(agentType string, state domain.MemoryPushState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.loadLocalConfig()
	if err != nil {
		return err
	}
	if cfg.PushState == nil {
		cfg.PushState = make(map[string]pushStateJSON)
	}
	cfg.PushState[agentType] = pushStateJSON{
		LastPushedAt:   state.LastPushedAt,
		LastPushedHash: state.LastPushedHash,
	}
	return s.saveLocalConfig(cfg)
}

// GetAllPushStates returns all stored push states as a map keyed by agent type.
func (s *FsStorage) GetAllPushStates() (map[string]domain.MemoryPushState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.loadLocalConfig()
	if err != nil {
		return nil, err
	}
	result := make(map[string]domain.MemoryPushState, len(cfg.PushState))
	for agentType, raw := range cfg.PushState {
		result[agentType] = domain.MemoryPushState{
			LastPushedAt:   raw.LastPushedAt,
			LastPushedHash: raw.LastPushedHash,
		}
	}
	return result, nil
}
