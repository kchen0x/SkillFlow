package main

import (
	agentapp "github.com/shinerio/skillflow/core/agentintegration/app"
	agentgateway "github.com/shinerio/skillflow/core/agentintegration/app/port/gateway"
	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
	agentgatewayinfra "github.com/shinerio/skillflow/core/agentintegration/infra/gateway"
	"github.com/shinerio/skillflow/core/config"
	"github.com/shinerio/skillflow/core/registry"
)

func registerAdapters() {
	agents := []string{"claude-code", "opencode", "codex", "gemini-cli", "openclaw"}
	for _, name := range agents {
		registry.RegisterAdapter(agentgatewayinfra.NewFilesystemAdapter(name, config.DefaultAgentPushDir(name)))
	}
}

func agentProfile(cfg config.AgentConfig) agentdomain.AgentProfile {
	return agentdomain.AgentProfile{
		Name:     cfg.Name,
		ScanDirs: append([]string(nil), cfg.ScanDirs...),
		PushDir:  cfg.PushDir,
		Enabled:  cfg.Enabled,
		Custom:   cfg.Custom,
	}
}

func agentProfiles(cfgs []config.AgentConfig) []agentdomain.AgentProfile {
	profiles := make([]agentdomain.AgentProfile, 0, len(cfgs))
	for _, cfg := range cfgs {
		profiles = append(profiles, agentProfile(cfg))
	}
	return profiles
}

func agentConfig(profile agentdomain.AgentProfile) config.AgentConfig {
	return config.AgentConfig{
		Name:     profile.Name,
		ScanDirs: append([]string(nil), profile.ScanDirs...),
		PushDir:  profile.PushDir,
		Enabled:  profile.Enabled,
		Custom:   profile.Custom,
	}
}

func agentConfigs(profiles []agentdomain.AgentProfile) []config.AgentConfig {
	cfgs := make([]config.AgentConfig, 0, len(profiles))
	for _, profile := range profiles {
		cfgs = append(cfgs, agentConfig(profile))
	}
	return cfgs
}

func resolveAgentGateway(profile agentdomain.AgentProfile) agentgateway.AgentGateway {
	if adapter, ok := registry.GetAdapter(profile.Name); ok {
		return adapter
	}
	return agentgatewayinfra.NewFilesystemAdapter(profile.Name, profile.PushDir)
}

func newAgentIntegrationService() *agentapp.Service {
	return agentapp.NewService(resolveAgentGateway)
}
