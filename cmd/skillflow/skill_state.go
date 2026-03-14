package main

import (
	"context"
	"os"
	"sort"
	"strings"

	"github.com/shinerio/skillflow/core/config"
	coregit "github.com/shinerio/skillflow/core/git"
	"github.com/shinerio/skillflow/core/install"
	"github.com/shinerio/skillflow/core/skill"
	"github.com/shinerio/skillflow/core/skillkey"
	agentsync "github.com/shinerio/skillflow/core/sync"
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

type AgentSkillCandidate struct {
	Name         string   `json:"name"`
	Path         string   `json:"path"`
	Source       string   `json:"source"`
	LogicalKey   string   `json:"logicalKey"`
	Installed    bool     `json:"installed"`
	Imported     bool     `json:"imported"`
	Updatable    bool     `json:"updatable"`
	Pushed       bool     `json:"pushed"`
	PushedAgents []string `json:"pushedAgents"`
}

type AgentSkillEntry struct {
	Name            string   `json:"name"`
	Path            string   `json:"path"`
	Source          string   `json:"source"`
	LogicalKey      string   `json:"logicalKey"`
	Installed       bool     `json:"installed"`
	Imported        bool     `json:"imported"`
	Updatable       bool     `json:"updatable"`
	Pushed          bool     `json:"pushed"`
	PushedAgents    []string `json:"pushedAgents"`
	SeenInAgentScan bool     `json:"seenInAgentScan"`
}

type resolvedSkillState struct {
	LogicalKey   string
	Source       string
	Installed    bool
	Imported     bool
	Updatable    bool
	Pushed       bool
	PushedAgents []string
}

type agentPresenceIndex struct {
	agentsByKey map[string][]string
}

func (a *App) installedIndex() ([]*skill.Skill, *skill.InstalledIndex, error) {
	installed, err := a.storage.ListAll()
	if err != nil {
		return nil, nil, err
	}
	return installed, skill.BuildInstalledIndex(installed), nil
}

func (a *App) buildInstalledSkillEntries(installed []*skill.Skill, presence *agentPresenceIndex) []InstalledSkillEntry {
	entries := make([]InstalledSkillEntry, 0, len(installed))
	for _, sk := range installed {
		if sk == nil {
			continue
		}
		logicalKey, err := skill.LogicalKey(sk)
		if err != nil {
			logicalKey = ""
		}
		pushedAgents := presence.agents(logicalKey)
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

func resolveGitHubCandidates(candidates []install.SkillCandidate, idx *skill.InstalledIndex, presence *agentPresenceIndex) []install.SkillCandidate {
	resolved := make([]install.SkillCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		state := resolveSkillState(candidate.Name, candidate.LogicalKey, idx, presence)
		candidate.LogicalKey = coalesceLogicalKey(candidate.LogicalKey, state.LogicalKey)
		candidate.Installed = state.Installed
		candidate.Updatable = state.Updatable
		candidate.Pushed = state.Pushed
		candidate.PushedAgents = state.PushedAgents
		resolved = append(resolved, candidate)
	}
	return resolved
}

func resolveStarSkills(skills []coregit.StarSkill, idx *skill.InstalledIndex, presence *agentPresenceIndex) []coregit.StarSkill {
	resolved := make([]coregit.StarSkill, 0, len(skills))
	for _, candidate := range skills {
		if strings.TrimSpace(candidate.LogicalKey) == "" {
			candidate.LogicalKey = skillkey.Git(candidate.Source, candidate.SubPath)
		}
		contentKey, err := skillkey.ContentFromDir(candidate.Path)
		if err != nil {
			contentKey = ""
		}
		state := resolveSkillState(candidate.Name, candidate.LogicalKey, idx, presence, contentKey)
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

func resolveAgentSkillCandidates(candidates []*skill.Skill, idx *skill.InstalledIndex, presence *agentPresenceIndex) []AgentSkillCandidate {
	byKey := map[string]AgentSkillCandidate{}
	for _, candidate := range candidates {
		if candidate == nil {
			continue
		}
		logicalKey, err := skillkey.ContentFromDir(candidate.Path)
		if err != nil {
			logicalKey = ""
		}
		state := resolveSkillState(candidate.Name, logicalKey, idx, presence)
		resolved := AgentSkillCandidate{
			Name:         candidate.Name,
			Path:         candidate.Path,
			Source:       state.Source,
			LogicalKey:   coalesceLogicalKey(state.LogicalKey, logicalKey),
			Installed:    state.Installed,
			Imported:     state.Imported,
			Updatable:    state.Updatable,
			Pushed:       state.Pushed,
			PushedAgents: state.PushedAgents,
		}
		groupKey := agentGroupKey(resolved.Name, resolved.LogicalKey, resolved.Path)
		if existing, ok := byKey[groupKey]; ok {
			byKey[groupKey] = mergeAgentCandidate(existing, resolved)
			continue
		}
		byKey[groupKey] = resolved
	}

	result := make([]AgentSkillCandidate, 0, len(byKey))
	for _, candidate := range byKey {
		result = append(result, candidate)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Name == result[j].Name {
			return result[i].Path < result[j].Path
		}
		return result[i].Name < result[j].Name
	})
	return result
}

func aggregateAgentSkillEntries(pushSkills, scanSkills []*skill.Skill, idx *skill.InstalledIndex, presence *agentPresenceIndex) []AgentSkillEntry {
	entries := map[string]AgentSkillEntry{}

	for _, candidate := range resolveAgentSkillCandidates(pushSkills, idx, presence) {
		groupKey := agentGroupKey(candidate.Name, candidate.LogicalKey, candidate.Path)
		entry := entries[groupKey]
		entry.Name = candidate.Name
		entry.Path = candidate.Path
		entry.Source = coalesceSource(candidate.Source, entry.Source)
		entry.LogicalKey = coalesceLogicalKey(candidate.LogicalKey, entry.LogicalKey)
		entry.Installed = entry.Installed || candidate.Installed
		entry.Imported = entry.Imported || candidate.Imported
		entry.Updatable = entry.Updatable || candidate.Updatable
		entry.Pushed = true
		entry.PushedAgents = mergeAgentLists(entry.PushedAgents, candidate.PushedAgents)
		entries[groupKey] = entry
	}

	for _, candidate := range resolveAgentSkillCandidates(scanSkills, idx, presence) {
		groupKey := agentGroupKey(candidate.Name, candidate.LogicalKey, candidate.Path)
		entry := entries[groupKey]
		if entry.Path == "" || !entry.Pushed {
			entry.Path = candidate.Path
		}
		if entry.Name == "" {
			entry.Name = candidate.Name
		}
		entry.Source = coalesceSource(candidate.Source, entry.Source)
		entry.LogicalKey = coalesceLogicalKey(candidate.LogicalKey, entry.LogicalKey)
		entry.Installed = entry.Installed || candidate.Installed
		entry.Imported = entry.Imported || candidate.Imported
		entry.Updatable = entry.Updatable || candidate.Updatable
		entry.PushedAgents = mergeAgentLists(entry.PushedAgents, candidate.PushedAgents)
		entry.SeenInAgentScan = true
		entries[groupKey] = entry
	}

	result := make([]AgentSkillEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, entry)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Name == result[j].Name {
			return result[i].Path < result[j].Path
		}
		return result[i].Name < result[j].Name
	})
	return result
}

func resolveAgentSkillSelection(candidates []AgentSkillCandidate, selectedPaths []string) []AgentSkillCandidate {
	selectedSet := make(map[string]struct{}, len(selectedPaths))
	for _, selectedPath := range selectedPaths {
		selectedSet[selectedPath] = struct{}{}
	}

	result := make([]AgentSkillCandidate, 0, len(selectedPaths))
	for _, candidate := range candidates {
		if _, ok := selectedSet[candidate.Path]; ok {
			result = append(result, candidate)
		}
	}
	return result
}

func scanAgentSkillsRaw(ctx context.Context, adapter agentsync.AgentAdapter, scanDirs []string, maxDepth int) ([]*skill.Skill, error) {
	var result []*skill.Skill
	seenPaths := map[string]struct{}{}
	for _, dir := range scanDirs {
		if dir == "" {
			continue
		}
		if _, err := os.Stat(dir); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		skills, err := pullAgentSkills(ctx, adapter, dir, maxDepth)
		if err != nil {
			return nil, err
		}
		for _, sk := range skills {
			if sk == nil || strings.TrimSpace(sk.Path) == "" {
				continue
			}
			if _, ok := seenPaths[sk.Path]; ok {
				continue
			}
			seenPaths[sk.Path] = struct{}{}
			result = append(result, sk)
		}
	}
	return result, nil
}

func (a *App) buildAgentPresenceIndex(idx *skill.InstalledIndex) *agentPresenceIndex {
	presence, _ := measureOperation(a, "build_agent_presence_index", func() (*agentPresenceIndex, error) {
		cfg, err := a.config.Load()
		if err != nil {
			a.logErrorf("build agent presence index failed: load config err=%v", err)
			return &agentPresenceIndex{agentsByKey: map[string][]string{}}, nil
		}

		snapshot, err := a.buildAgentPresenceSnapshot(cfg, idx)
		if err != nil {
			a.logErrorf("build agent presence index failed: build snapshot err=%v", err)
			return &agentPresenceIndex{agentsByKey: map[string][]string{}}, nil
		}

		return &agentPresenceIndex{agentsByKey: snapshot.AgentsByKey}, nil
	})
	return presence
}

func (a *App) indexAgentPushPresence(presence *agentPresenceIndex, idx *skill.InstalledIndex, agent config.AgentConfig) {
	if presence == nil || strings.TrimSpace(agent.PushDir) == "" {
		return
	}
	if _, err := os.Stat(agent.PushDir); err != nil {
		if !os.IsNotExist(err) {
			a.logErrorf("build agent presence index failed: agent=%s pushDir=%s err=%v", agent.Name, agent.PushDir, err)
		}
		return
	}

	pushed, err := pullAgentSkills(a.ctx, getAdapter(agent), agent.PushDir, a.repoScanMaxDepth())
	if err != nil {
		a.logErrorf("build agent presence index failed: agent=%s pushDir=%s err=%v", agent.Name, agent.PushDir, err)
		return
	}

	for _, candidate := range pushed {
		if candidate == nil {
			continue
		}
		presence.add(agent.Name, agentPresenceKeys(candidate, idx)...)
	}
}

func resolveSkillState(name, logicalKey string, idx *skill.InstalledIndex, presence *agentPresenceIndex, aliasKeys ...string) resolvedSkillState {
	status := idx.Resolve(name, logicalKey)
	resolvedKey := coalesceLogicalKey(status.LogicalKey, logicalKey)
	lookupKeys := []string{resolvedKey}
	for _, aliasKey := range aliasKeys {
		if strings.TrimSpace(aliasKey) != "" {
			lookupKeys = append(lookupKeys, aliasKey)
		}
	}
	pushedAgents := presence.agents(lookupKeys...)
	return resolvedSkillState{
		LogicalKey:   coalesceLogicalKey(resolvedKey, logicalKey),
		Source:       status.Source,
		Installed:    status.Installed,
		Imported:     status.Imported,
		Updatable:    status.Updatable,
		Pushed:       len(pushedAgents) > 0,
		PushedAgents: pushedAgents,
	}
}

func agentPresenceKeys(candidate *skill.Skill, idx *skill.InstalledIndex) []string {
	if candidate == nil {
		return nil
	}
	contentKey, err := skillkey.ContentFromDir(candidate.Path)
	if err != nil {
		contentKey = ""
	}
	status := idx.Resolve(candidate.Name, contentKey)
	return compactKeys(status.LogicalKey, contentKey)
}

func mergeAgentCandidate(existing, incoming AgentSkillCandidate) AgentSkillCandidate {
	if existing.Path == "" {
		existing.Path = incoming.Path
	}
	if existing.LogicalKey == "" {
		existing.LogicalKey = incoming.LogicalKey
	}
	if existing.Source == "" {
		existing.Source = incoming.Source
	}
	existing.Installed = existing.Installed || incoming.Installed
	existing.Imported = existing.Imported || incoming.Imported
	existing.Updatable = existing.Updatable || incoming.Updatable
	existing.Pushed = existing.Pushed || incoming.Pushed
	existing.PushedAgents = mergeAgentLists(existing.PushedAgents, incoming.PushedAgents)
	return existing
}

func (p *agentPresenceIndex) add(agentName string, keys ...string) {
	if p == nil || strings.TrimSpace(agentName) == "" {
		return
	}
	for _, key := range compactKeys(keys...) {
		p.agentsByKey[key] = appendUniqueAgent(p.agentsByKey[key], agentName)
	}
}

func (p *agentPresenceIndex) agents(keys ...string) []string {
	if p == nil {
		return nil
	}
	var merged []string
	for _, key := range compactKeys(keys...) {
		merged = mergeAgentLists(merged, p.agentsByKey[key])
	}
	return merged
}

func mergeAgentLists(primary, secondary []string) []string {
	merged := append([]string{}, primary...)
	for _, agentName := range secondary {
		merged = appendUniqueAgent(merged, agentName)
	}
	return merged
}

func appendUniqueAgent(agents []string, agentName string) []string {
	for _, existing := range agents {
		if existing == agentName {
			return agents
		}
	}
	return append(agents, agentName)
}

func compactKeys(keys ...string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, key)
	}
	return result
}

func coalesceLogicalKey(primary, secondary string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return secondary
}

func coalesceSource(primary, secondary string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return secondary
}

func agentGroupKey(name, logicalKey, path string) string {
	if strings.TrimSpace(logicalKey) != "" {
		return "logical:" + logicalKey
	}
	if trimmedName := strings.ToLower(strings.TrimSpace(name)); trimmedName != "" {
		return "fallback:" + trimmedName
	}
	return "path:" + path
}
