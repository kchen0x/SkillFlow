package main

import "github.com/shinerio/skillflow/core/orchestration"

type cloudRestoreState = orchestration.RestoreState

func (a *App) captureCloudRestoreState() cloudRestoreState {
	state, err := a.newOrchestrationService().CaptureRestoreState(a.ctx)
	if err != nil {
		a.logErrorf("capture restore state failed: %v", err)
		return cloudRestoreState{
			InstalledSkills: map[string]orchestration.RestoreSkillSnapshot{},
			StarredRepoURLs: map[string]struct{}{},
		}
	}
	return state
}

func (a *App) handleRestoredCloudState(before cloudRestoreState, source string) error {
	a.reloadStateFromDisk()

	cfg, err := a.config.Load()
	if err != nil {
		a.logErrorf("restore compensation failed: source=%s load config failed: %v", source, err)
		return err
	}

	_, err = a.newOrchestrationService().CompensateRestore(a.ctx, orchestration.RestoreCompensationCommand{
		Before:             before,
		Source:             source,
		AgentProfiles:      cfg.Agents,
		AutoPushAgentNames: cfg.AutoPushAgents,
		ProxyURL:           a.gitProxyURL(),
	})
	return err
}
