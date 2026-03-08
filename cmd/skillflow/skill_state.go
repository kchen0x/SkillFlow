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
	toolsync "github.com/shinerio/skillflow/core/sync"
)

type InstalledSkillEntry struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Category    string   `json:"category"`
	Source      string   `json:"source"`
	SourceSHA   string   `json:"sourceSha"`
	LatestSHA   string   `json:"latestSha"`
	Updatable   bool     `json:"updatable"`
	Pushed      bool     `json:"pushed"`
	PushedTools []string `json:"pushedTools"`
}

type ToolSkillCandidate struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	LogicalKey  string   `json:"logicalKey"`
	Installed   bool     `json:"installed"`
	Imported    bool     `json:"imported"`
	Updatable   bool     `json:"updatable"`
	Pushed      bool     `json:"pushed"`
	PushedTools []string `json:"pushedTools"`
}

type ToolSkillEntry struct {
	Name           string   `json:"name"`
	Path           string   `json:"path"`
	LogicalKey     string   `json:"logicalKey"`
	Installed      bool     `json:"installed"`
	Imported       bool     `json:"imported"`
	Updatable      bool     `json:"updatable"`
	Pushed         bool     `json:"pushed"`
	PushedTools    []string `json:"pushedTools"`
	SeenInToolScan bool     `json:"seenInToolScan"`
}

type resolvedSkillState struct {
	LogicalKey  string
	Installed   bool
	Imported    bool
	Updatable   bool
	Pushed      bool
	PushedTools []string
}

type toolPresenceIndex struct {
	toolsByKey map[string][]string
}

func (a *App) installedIndex() ([]*skill.Skill, *skill.InstalledIndex, error) {
	installed, err := a.storage.ListAll()
	if err != nil {
		return nil, nil, err
	}
	return installed, skill.BuildInstalledIndex(installed), nil
}

func (a *App) buildInstalledSkillEntries(installed []*skill.Skill, presence *toolPresenceIndex) []InstalledSkillEntry {
	entries := make([]InstalledSkillEntry, 0, len(installed))
	for _, sk := range installed {
		if sk == nil {
			continue
		}
		logicalKey, err := skill.LogicalKey(sk)
		if err != nil {
			logicalKey = ""
		}
		pushedTools := presence.tools(logicalKey)
		entries = append(entries, InstalledSkillEntry{
			ID:          sk.ID,
			Name:        sk.Name,
			Path:        sk.Path,
			Category:    sk.Category,
			Source:      string(sk.Source),
			SourceSHA:   sk.SourceSHA,
			LatestSHA:   sk.LatestSHA,
			Updatable:   sk.HasUpdate(),
			Pushed:      len(pushedTools) > 0,
			PushedTools: pushedTools,
		})
	}
	return entries
}

func resolveGitHubCandidates(candidates []install.SkillCandidate, idx *skill.InstalledIndex, presence *toolPresenceIndex) []install.SkillCandidate {
	resolved := make([]install.SkillCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		state := resolveSkillState(candidate.Name, candidate.LogicalKey, idx, presence)
		candidate.LogicalKey = coalesceLogicalKey(candidate.LogicalKey, state.LogicalKey)
		candidate.Installed = state.Installed
		candidate.Updatable = state.Updatable
		candidate.Pushed = state.Pushed
		candidate.PushedTools = state.PushedTools
		resolved = append(resolved, candidate)
	}
	return resolved
}

func resolveStarSkills(skills []coregit.StarSkill, idx *skill.InstalledIndex, presence *toolPresenceIndex) []coregit.StarSkill {
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
		candidate.PushedTools = state.PushedTools
		resolved = append(resolved, candidate)
	}
	return resolved
}

func resolveToolSkillCandidates(candidates []*skill.Skill, idx *skill.InstalledIndex, presence *toolPresenceIndex) []ToolSkillCandidate {
	byKey := map[string]ToolSkillCandidate{}
	for _, candidate := range candidates {
		if candidate == nil {
			continue
		}
		logicalKey, err := skillkey.ContentFromDir(candidate.Path)
		if err != nil {
			logicalKey = ""
		}
		state := resolveSkillState(candidate.Name, logicalKey, idx, presence)
		resolved := ToolSkillCandidate{
			Name:        candidate.Name,
			Path:        candidate.Path,
			LogicalKey:  coalesceLogicalKey(state.LogicalKey, logicalKey),
			Installed:   state.Installed,
			Imported:    state.Imported,
			Updatable:   state.Updatable,
			Pushed:      state.Pushed,
			PushedTools: state.PushedTools,
		}
		groupKey := toolGroupKey(resolved.Name, resolved.LogicalKey, resolved.Path)
		if existing, ok := byKey[groupKey]; ok {
			byKey[groupKey] = mergeToolCandidate(existing, resolved)
			continue
		}
		byKey[groupKey] = resolved
	}

	result := make([]ToolSkillCandidate, 0, len(byKey))
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

func aggregateToolSkillEntries(pushSkills, scanSkills []*skill.Skill, idx *skill.InstalledIndex, presence *toolPresenceIndex) []ToolSkillEntry {
	entries := map[string]ToolSkillEntry{}

	for _, candidate := range resolveToolSkillCandidates(pushSkills, idx, presence) {
		groupKey := toolGroupKey(candidate.Name, candidate.LogicalKey, candidate.Path)
		entry := entries[groupKey]
		entry.Name = candidate.Name
		entry.Path = candidate.Path
		entry.LogicalKey = coalesceLogicalKey(candidate.LogicalKey, entry.LogicalKey)
		entry.Installed = entry.Installed || candidate.Installed
		entry.Imported = entry.Imported || candidate.Imported
		entry.Updatable = entry.Updatable || candidate.Updatable
		entry.Pushed = true
		entry.PushedTools = mergeToolLists(entry.PushedTools, candidate.PushedTools)
		entries[groupKey] = entry
	}

	for _, candidate := range resolveToolSkillCandidates(scanSkills, idx, presence) {
		groupKey := toolGroupKey(candidate.Name, candidate.LogicalKey, candidate.Path)
		entry := entries[groupKey]
		if entry.Path == "" || !entry.Pushed {
			entry.Path = candidate.Path
		}
		if entry.Name == "" {
			entry.Name = candidate.Name
		}
		entry.LogicalKey = coalesceLogicalKey(candidate.LogicalKey, entry.LogicalKey)
		entry.Installed = entry.Installed || candidate.Installed
		entry.Imported = entry.Imported || candidate.Imported
		entry.Updatable = entry.Updatable || candidate.Updatable
		entry.PushedTools = mergeToolLists(entry.PushedTools, candidate.PushedTools)
		entry.SeenInToolScan = true
		entries[groupKey] = entry
	}

	result := make([]ToolSkillEntry, 0, len(entries))
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

func resolveToolSkillSelection(candidates []ToolSkillCandidate, selectedPaths []string) []ToolSkillCandidate {
	selectedSet := make(map[string]struct{}, len(selectedPaths))
	for _, selectedPath := range selectedPaths {
		selectedSet[selectedPath] = struct{}{}
	}

	result := make([]ToolSkillCandidate, 0, len(selectedPaths))
	for _, candidate := range candidates {
		if _, ok := selectedSet[candidate.Path]; ok {
			result = append(result, candidate)
		}
	}
	return result
}

func scanToolSkillsRaw(ctx context.Context, adapter toolsync.ToolAdapter, scanDirs []string, maxDepth int) ([]*skill.Skill, error) {
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
		skills, err := pullToolSkills(ctx, adapter, dir, maxDepth)
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

func (a *App) buildToolPresenceIndex(idx *skill.InstalledIndex) *toolPresenceIndex {
	presence := &toolPresenceIndex{toolsByKey: map[string][]string{}}

	cfg, err := a.config.Load()
	if err != nil {
		a.logErrorf("build tool presence index failed: load config err=%v", err)
		return presence
	}

	for _, tool := range cfg.Tools {
		a.indexToolPushPresence(presence, idx, tool)
	}

	return presence
}

func (a *App) indexToolPushPresence(presence *toolPresenceIndex, idx *skill.InstalledIndex, tool config.ToolConfig) {
	if presence == nil || strings.TrimSpace(tool.PushDir) == "" {
		return
	}
	if _, err := os.Stat(tool.PushDir); err != nil {
		if !os.IsNotExist(err) {
			a.logErrorf("build tool presence index failed: tool=%s pushDir=%s err=%v", tool.Name, tool.PushDir, err)
		}
		return
	}

	pushed, err := pullToolSkills(a.ctx, getAdapter(tool), tool.PushDir, a.repoScanMaxDepth())
	if err != nil {
		a.logErrorf("build tool presence index failed: tool=%s pushDir=%s err=%v", tool.Name, tool.PushDir, err)
		return
	}

	for _, candidate := range pushed {
		if candidate == nil {
			continue
		}
		presence.add(tool.Name, toolPresenceKeys(candidate, idx)...)
	}
}

func resolveSkillState(name, logicalKey string, idx *skill.InstalledIndex, presence *toolPresenceIndex, aliasKeys ...string) resolvedSkillState {
	status := idx.Resolve(name, logicalKey)
	resolvedKey := coalesceLogicalKey(status.LogicalKey, logicalKey)
	lookupKeys := []string{resolvedKey}
	for _, aliasKey := range aliasKeys {
		if strings.TrimSpace(aliasKey) != "" {
			lookupKeys = append(lookupKeys, aliasKey)
		}
	}
	pushedTools := presence.tools(lookupKeys...)
	return resolvedSkillState{
		LogicalKey:  coalesceLogicalKey(resolvedKey, logicalKey),
		Installed:   status.Installed,
		Imported:    status.Imported,
		Updatable:   status.Updatable,
		Pushed:      len(pushedTools) > 0,
		PushedTools: pushedTools,
	}
}

func toolPresenceKeys(candidate *skill.Skill, idx *skill.InstalledIndex) []string {
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

func mergeToolCandidate(existing, incoming ToolSkillCandidate) ToolSkillCandidate {
	if existing.Path == "" {
		existing.Path = incoming.Path
	}
	if existing.LogicalKey == "" {
		existing.LogicalKey = incoming.LogicalKey
	}
	existing.Installed = existing.Installed || incoming.Installed
	existing.Imported = existing.Imported || incoming.Imported
	existing.Updatable = existing.Updatable || incoming.Updatable
	existing.Pushed = existing.Pushed || incoming.Pushed
	existing.PushedTools = mergeToolLists(existing.PushedTools, incoming.PushedTools)
	return existing
}

func (p *toolPresenceIndex) add(toolName string, keys ...string) {
	if p == nil || strings.TrimSpace(toolName) == "" {
		return
	}
	for _, key := range compactKeys(keys...) {
		p.toolsByKey[key] = appendUniqueTool(p.toolsByKey[key], toolName)
	}
}

func (p *toolPresenceIndex) tools(keys ...string) []string {
	if p == nil {
		return nil
	}
	var merged []string
	for _, key := range compactKeys(keys...) {
		merged = mergeToolLists(merged, p.toolsByKey[key])
	}
	return merged
}

func mergeToolLists(primary, secondary []string) []string {
	merged := append([]string{}, primary...)
	for _, toolName := range secondary {
		merged = appendUniqueTool(merged, toolName)
	}
	return merged
}

func appendUniqueTool(tools []string, toolName string) []string {
	for _, existing := range tools {
		if existing == toolName {
			return tools
		}
	}
	return append(tools, toolName)
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

func toolGroupKey(name, logicalKey, path string) string {
	if strings.TrimSpace(logicalKey) != "" {
		return "logical:" + logicalKey
	}
	if trimmedName := strings.ToLower(strings.TrimSpace(name)); trimmedName != "" {
		return "fallback:" + trimmedName
	}
	return "path:" + path
}
