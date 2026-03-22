package main

import (
	"fmt"

	agentapp "github.com/shinerio/skillflow/core/agentintegration/app"
	"github.com/shinerio/skillflow/core/readmodel/agentmemory"
)

type AgentMemoryPreviewDTO struct {
	AgentName      string                `json:"agentName"`
	MemoryPath     string                `json:"memoryPath"`
	RulesDir       string                `json:"rulesDir"`
	MainExists     bool                  `json:"mainExists"`
	MainContent    string                `json:"mainContent"`
	RulesDirExists bool                  `json:"rulesDirExists"`
	Rules          []*AgentMemoryRuleDTO `json:"rules"`
}

type AgentMemoryRuleDTO struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Content string `json:"content"`
	Managed bool   `json:"managed"`
}

func (a *App) GetAgentMemoryPreview(agentName string) (*AgentMemoryPreviewDTO, error) {
	a.logInfof("agent memory preview load started: agent=%s", agentName)

	cfg, err := a.config.Load()
	if err != nil {
		a.logErrorf("agent memory preview load failed: agent=%s load config failed: %v", agentName, err)
		return nil, err
	}
	profile, ok := agentapp.FindProfile(cfg.Agents, agentName)
	if !ok {
		err := fmt.Errorf("agent %s not found", agentName)
		a.logErrorf("agent memory preview load failed: agent=%s err=%v", agentName, err)
		return nil, err
	}

	preview, err := agentmemory.LoadPreview(profile)
	if err != nil {
		a.logErrorf("agent memory preview load failed: agent=%s err=%v", agentName, err)
		return nil, err
	}

	dto := &AgentMemoryPreviewDTO{
		AgentName:      preview.AgentName,
		MemoryPath:     preview.MemoryPath,
		RulesDir:       preview.RulesDir,
		MainExists:     preview.MainExists,
		MainContent:    preview.MainContent,
		RulesDirExists: preview.RulesDirExists,
		Rules:          make([]*AgentMemoryRuleDTO, 0, len(preview.Rules)),
	}
	for _, rule := range preview.Rules {
		dto.Rules = append(dto.Rules, &AgentMemoryRuleDTO{
			Name:    rule.Name,
			Path:    rule.Path,
			Content: rule.Content,
			Managed: rule.Managed,
		})
	}

	a.logInfof("agent memory preview load completed: agent=%s rules=%d mainExists=%t rulesDirExists=%t", agentName, len(dto.Rules), dto.MainExists, dto.RulesDirExists)
	return dto, nil
}
