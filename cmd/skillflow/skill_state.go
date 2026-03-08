package main

import (
	"context"
	"os"
	"sort"
	"strings"

	coregit "github.com/shinerio/skillflow/core/git"
	"github.com/shinerio/skillflow/core/install"
	"github.com/shinerio/skillflow/core/skill"
	"github.com/shinerio/skillflow/core/skillkey"
	toolsync "github.com/shinerio/skillflow/core/sync"
)

type ToolSkillCandidate struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	LogicalKey string `json:"logicalKey"`
	Installed  bool   `json:"installed"`
	Imported   bool   `json:"imported"`
	Updatable  bool   `json:"updatable"`
}

type ToolSkillEntry struct {
	Name           string `json:"name"`
	Path           string `json:"path"`
	LogicalKey     string `json:"logicalKey"`
	Installed      bool   `json:"installed"`
	Imported       bool   `json:"imported"`
	Updatable      bool   `json:"updatable"`
	Pushed         bool   `json:"pushed"`
	SeenInToolScan bool   `json:"seenInToolScan"`
}

func (a *App) installedIndex() ([]*skill.Skill, *skill.InstalledIndex, error) {
	installed, err := a.storage.ListAll()
	if err != nil {
		return nil, nil, err
	}
	return installed, skill.BuildInstalledIndex(installed), nil
}

func resolveGitHubCandidates(candidates []install.SkillCandidate, idx *skill.InstalledIndex) []install.SkillCandidate {
	resolved := make([]install.SkillCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		status := idx.Resolve(candidate.Name, candidate.LogicalKey)
		candidate.Installed = status.Installed
		candidate.Updatable = status.Updatable
		resolved = append(resolved, candidate)
	}
	return resolved
}

func resolveStarSkills(skills []coregit.StarSkill, idx *skill.InstalledIndex) []coregit.StarSkill {
	resolved := make([]coregit.StarSkill, 0, len(skills))
	for _, candidate := range skills {
		if strings.TrimSpace(candidate.LogicalKey) == "" {
			candidate.LogicalKey = skillkey.Git(candidate.Source, candidate.SubPath)
		}
		status := idx.Resolve(candidate.Name, candidate.LogicalKey)
		candidate.Installed = status.Installed
		candidate.Imported = status.Imported
		candidate.Updatable = status.Updatable
		resolved = append(resolved, candidate)
	}
	return resolved
}

func resolveToolSkillCandidates(candidates []*skill.Skill, idx *skill.InstalledIndex) []ToolSkillCandidate {
	byKey := map[string]ToolSkillCandidate{}
	for _, candidate := range candidates {
		if candidate == nil {
			continue
		}
		logicalKey, err := skillkey.ContentFromDir(candidate.Path)
		if err != nil {
			logicalKey = ""
		}
		status := idx.Resolve(candidate.Name, logicalKey)
		resolved := ToolSkillCandidate{
			Name:       candidate.Name,
			Path:       candidate.Path,
			LogicalKey: coalesceLogicalKey(logicalKey, status.LogicalKey),
			Installed:  status.Installed,
			Imported:   status.Imported,
			Updatable:  status.Updatable,
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

func aggregateToolSkillEntries(pushSkills, scanSkills []*skill.Skill, idx *skill.InstalledIndex) []ToolSkillEntry {
	entries := map[string]ToolSkillEntry{}

	for _, candidate := range resolveToolSkillCandidates(pushSkills, idx) {
		groupKey := toolGroupKey(candidate.Name, candidate.LogicalKey, candidate.Path)
		entry := entries[groupKey]
		entry.Name = candidate.Name
		entry.Path = candidate.Path
		entry.LogicalKey = coalesceLogicalKey(candidate.LogicalKey, entry.LogicalKey)
		entry.Installed = entry.Installed || candidate.Installed
		entry.Imported = entry.Imported || candidate.Imported
		entry.Updatable = entry.Updatable || candidate.Updatable
		entry.Pushed = true
		entries[groupKey] = entry
	}

	for _, candidate := range resolveToolSkillCandidates(scanSkills, idx) {
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
	return existing
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
