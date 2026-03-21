package main

import (
	"strings"

	agentapp "github.com/shinerio/skillflow/core/agentintegration/app"
	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
	coregit "github.com/shinerio/skillflow/core/git"
	skillquery "github.com/shinerio/skillflow/core/skillcatalog/app/query"
	skilldomain "github.com/shinerio/skillflow/core/skillcatalog/domain"
	"github.com/shinerio/skillflow/core/skillkey"
)

type InstalledSkillEntry struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Path         string   `json:"path"`
	Category     string   `json:"category"`
	Source       string   `json:"source"`
	SourceSHA    string   `json:"sourceSha"`
	LatestSHA    string   `json:"latestSha"`
	Updatable    bool     `json:"updatable"`
	Pushed       bool     `json:"pushed"`
	PushedAgents []string `json:"pushedAgents"`
}

func (a *App) installedIndex() ([]*skilldomain.InstalledSkill, *skillquery.InstalledIndex, error) {
	installed, err := a.storage.ListAll()
	if err != nil {
		return nil, nil, err
	}
	return installed, skillquery.BuildInstalledIndex(installed), nil
}

func (a *App) buildInstalledSkillEntries(installed []*skilldomain.InstalledSkill, presence *agentdomain.AgentPresenceIndex) []InstalledSkillEntry {
	entries := make([]InstalledSkillEntry, 0, len(installed))
	for _, sk := range installed {
		if sk == nil {
			continue
		}
		logicalKey, err := skilldomain.LogicalKey(sk)
		if err != nil {
			logicalKey = ""
		}
		pushedAgents := presence.Agents(logicalKey)
		entries = append(entries, InstalledSkillEntry{
			ID:           sk.ID,
			Name:         sk.Name,
			Path:         sk.Path,
			Category:     sk.Category,
			Source:       string(sk.Source),
			SourceSHA:    sk.SourceSHA,
			LatestSHA:    sk.LatestSHA,
			Updatable:    sk.HasUpdate(),
			Pushed:       len(pushedAgents) > 0,
			PushedAgents: pushedAgents,
		})
	}
	return entries
}

func resolveStarSkills(skills []coregit.StarSkill, idx *skillquery.InstalledIndex, presence *agentdomain.AgentPresenceIndex) []coregit.StarSkill {
	resolved := make([]coregit.StarSkill, 0, len(skills))
	for _, candidate := range skills {
		if strings.TrimSpace(candidate.LogicalKey) == "" {
			candidate.LogicalKey = skillkey.Git(candidate.Source, candidate.SubPath)
		}
		contentKey, err := skillkey.ContentFromDir(candidate.Path)
		if err != nil {
			contentKey = ""
		}
		state := agentapp.ResolveSkillStatus(candidate.Name, candidate.LogicalKey, idx, presence, contentKey)
		candidate.LogicalKey = coalesceLogicalKey(candidate.LogicalKey, state.LogicalKey)
		candidate.Installed = state.Installed
		candidate.Imported = state.Imported
		candidate.Updatable = state.Updatable
		candidate.Pushed = state.Pushed
		candidate.PushedAgents = state.PushedAgents
		resolved = append(resolved, candidate)
	}
	return resolved
}

func (a *App) buildAgentPresenceIndex(idx *skillquery.InstalledIndex) *agentdomain.AgentPresenceIndex {
	presence, _ := measureOperation(a, "build_agent_presence_index", func() (*agentdomain.AgentPresenceIndex, error) {
		cfg, err := a.config.Load()
		if err != nil {
			a.logErrorf("build agent presence index failed: load config err=%v", err)
			return agentdomain.NewAgentPresenceIndex(), nil
		}

		snapshot, err := a.buildAgentPresenceSnapshot(cfg, idx)
		if err != nil {
			a.logErrorf("build agent presence index failed: build snapshot err=%v", err)
			return agentdomain.NewAgentPresenceIndex(), nil
		}
		if len(snapshot.AgentsByKey) > 0 {
			presence := agentdomain.NewAgentPresenceIndex()
			for key, agentNames := range snapshot.AgentsByKey {
				for _, agentName := range agentNames {
					presence.Add(agentName, key)
				}
			}
			return presence, nil
		}

		service := newAgentIntegrationService()
		presence, err := service.BuildPresenceIndex(a.ctx, agentProfiles(cfg.Agents), idx, a.repoScanMaxDepth())
		if err != nil {
			a.logErrorf("build agent presence index failed: build presence err=%v", err)
			return agentdomain.NewAgentPresenceIndex(), nil
		}
		return presence, nil
	})
	return presence
}

func coalesceLogicalKey(primary, secondary string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return secondary
}
