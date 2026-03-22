package app

import (
	"strings"
	"testing"
	"time"

	gatewayport "github.com/shinerio/skillflow/core/memorycatalog/app/port/gateway"
	"github.com/shinerio/skillflow/core/memorycatalog/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPushSelectionToAgentWritesOnlySelectedModulesAndRemovesOthers(t *testing.T) {
	storage := &testMemoryStorage{
		main: &domain.MainMemory{Content: "Main memory body", UpdatedAt: time.Now()},
		modules: map[string]*domain.ModuleMemory{
			"style":   {Name: "style", Content: "Style rules", UpdatedAt: time.Now()},
			"testing": {Name: "testing", Content: "Always test", UpdatedAt: time.Now()},
		},
		pushState: make(map[string]domain.MemoryPushState),
	}
	pusher := &recordingPusher{}
	service := NewPushService(storage, testAgentConfigGateway{
		agents: map[string]gatewayport.AgentMemoryConfig{
			"codex": {
				AgentType:  "codex",
				MemoryPath: "/tmp/AGENTS.md",
				RulesDir:   "/tmp/rules",
			},
		},
	}, func(agentType string) (gatewayport.AgentMemoryPusher, bool) {
		pusher.managedModules = []string{"style", "testing"}
		return pusher, true
	})

	require.NoError(t, service.PushSelectionToAgent("codex", []string{"testing"}, domain.PushModeMerge))

	assert.Equal(t, []string{"testing"}, pusher.pushedModules)
	assert.Equal(t, []string{"style"}, pusher.removedModules)
	assert.Equal(t, domain.PushModeMerge, pusher.mainMode)
	assert.Equal(t, strings.TrimSpace(`
<skillflow-managed>
Main memory body
</skillflow-managed>

<skillflow-module>
Please be sure to load all module memories below.
[testing](rules/sf-testing.md)
</skillflow-module>`), pusher.mainContent)
	assert.NotEmpty(t, storage.pushState["codex"].LastPushedHash)
}

func TestPushSelectionToAgentLeavesAgentPendingWhenSelectionIsPartial(t *testing.T) {
	storage := &testMemoryStorage{
		main: &domain.MainMemory{Content: "Main memory", UpdatedAt: time.Now()},
		modules: map[string]*domain.ModuleMemory{
			"style":   {Name: "style", Content: "Style rules", UpdatedAt: time.Now()},
			"testing": {Name: "testing", Content: "Always test", UpdatedAt: time.Now()},
		},
		pushState: make(map[string]domain.MemoryPushState),
	}
	service := NewPushService(storage, testAgentConfigGateway{
		agents: map[string]gatewayport.AgentMemoryConfig{
			"codex": {
				AgentType:  "codex",
				MemoryPath: "/tmp/AGENTS.md",
				RulesDir:   "/tmp/rules",
			},
		},
	}, func(agentType string) (gatewayport.AgentMemoryPusher, bool) {
		return &recordingPusher{}, true
	})

	require.NoError(t, service.PushSelectionToAgent("codex", []string{"testing"}, domain.PushModeTakeover))

	status, err := service.GetPushStatus("codex")
	require.NoError(t, err)
	assert.Equal(t, "pendingPush", status)
}

type testMemoryStorage struct {
	main      *domain.MainMemory
	modules   map[string]*domain.ModuleMemory
	pushState map[string]domain.MemoryPushState
}

func (s *testMemoryStorage) GetMainMemory() (*domain.MainMemory, error) {
	return s.main, nil
}

func (s *testMemoryStorage) SaveMainMemory(content string) (*domain.MainMemory, error) {
	s.main = &domain.MainMemory{Content: content, UpdatedAt: time.Now()}
	return s.main, nil
}

func (s *testMemoryStorage) ListModules() ([]*domain.ModuleMemory, error) {
	result := make([]*domain.ModuleMemory, 0, len(s.modules))
	for _, module := range s.modules {
		result = append(result, module)
	}
	return result, nil
}

func (s *testMemoryStorage) GetModule(name string) (*domain.ModuleMemory, error) {
	return s.modules[name], nil
}

func (s *testMemoryStorage) CreateModule(name, content string) (*domain.ModuleMemory, error) {
	panic("unexpected call")
}

func (s *testMemoryStorage) SaveModule(name, content string) (*domain.ModuleMemory, error) {
	panic("unexpected call")
}

func (s *testMemoryStorage) DeleteModule(name string) error {
	panic("unexpected call")
}

func (s *testMemoryStorage) GetPushConfig(agentType string) (domain.MemoryPushConfig, error) {
	return domain.MemoryPushConfig{}, nil
}

func (s *testMemoryStorage) SavePushConfig(cfg domain.MemoryPushConfig) error {
	panic("unexpected call")
}

func (s *testMemoryStorage) GetAllPushConfigs() ([]domain.MemoryPushConfig, error) {
	return nil, nil
}

func (s *testMemoryStorage) GetModulePushTargets(moduleName string) (domain.ModulePushTargets, error) {
	return domain.ModulePushTargets{}, nil
}

func (s *testMemoryStorage) SaveModulePushTargets(targets domain.ModulePushTargets) error {
	panic("unexpected call")
}

func (s *testMemoryStorage) GetAllModulePushTargets() ([]domain.ModulePushTargets, error) {
	return nil, nil
}

func (s *testMemoryStorage) DeleteModulePushTargets(moduleName string) error {
	panic("unexpected call")
}

func (s *testMemoryStorage) GetPushState(agentType string) (domain.MemoryPushState, error) {
	return s.pushState[agentType], nil
}

func (s *testMemoryStorage) SavePushState(agentType string, state domain.MemoryPushState) error {
	s.pushState[agentType] = state
	return nil
}

func (s *testMemoryStorage) GetAllPushStates() (map[string]domain.MemoryPushState, error) {
	return s.pushState, nil
}

type testAgentConfigGateway struct {
	agents map[string]gatewayport.AgentMemoryConfig
}

func (g testAgentConfigGateway) ListEnabledAgents() ([]gatewayport.AgentMemoryConfig, error) {
	result := make([]gatewayport.AgentMemoryConfig, 0, len(g.agents))
	for _, agent := range g.agents {
		result = append(result, agent)
	}
	return result, nil
}

func (g testAgentConfigGateway) GetAgent(agentType string) (gatewayport.AgentMemoryConfig, bool, error) {
	agent, ok := g.agents[agentType]
	return agent, ok, nil
}

type recordingPusher struct {
	mainContent    string
	mainMode       domain.PushMode
	pushedModules  []string
	removedModules []string
	managedModules []string
}

func (p *recordingPusher) PushMainMemory(content string, mode domain.PushMode, agentMemoryPath string) error {
	p.mainContent = content
	p.mainMode = mode
	return nil
}

func (p *recordingPusher) PushModuleMemory(module *domain.ModuleMemory, agentRulesDir string) error {
	p.pushedModules = append(p.pushedModules, module.Name)
	return nil
}

func (p *recordingPusher) ListManagedModuleNames(agentRulesDir string) ([]string, error) {
	return append([]string(nil), p.managedModules...), nil
}

func (p *recordingPusher) RemoveModuleMemory(moduleName string, agentRulesDir string) error {
	p.removedModules = append(p.removedModules, moduleName)
	return nil
}

func (p *recordingPusher) BuildRulesIndex(modules []*domain.ModuleMemory, agentMemoryPath string, agentRulesDir string) gatewayport.RulesIndex {
	entries := make([]string, 0, len(modules))
	for _, module := range modules {
		entries = append(entries, "["+module.Name+"](rules/sf-"+module.Name+".md)")
	}
	return gatewayport.RulesIndex{
		Entries: entries,
	}
}

func (p *recordingPusher) RepairManagedBlock(agentMemoryPath string) error {
	return nil
}

func TestComposeManagedMemoryAddsModuleLoadInstructionAsFirstLine(t *testing.T) {
	content := composeManagedMemory("Main memory body", gatewayport.RulesIndex{
		Entries: []string{"[testing](rules/sf-testing.md)"},
	})

	assert.Equal(t, strings.TrimSpace(`
<skillflow-managed>
Main memory body
</skillflow-managed>

<skillflow-module>
Please be sure to load all module memories below.
[testing](rules/sf-testing.md)
</skillflow-module>`), content)
}
