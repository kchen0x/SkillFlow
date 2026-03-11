package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/shinerio/skillflow/core/config"
	coregit "github.com/shinerio/skillflow/core/git"
	"github.com/shinerio/skillflow/core/skill"
	"github.com/shinerio/skillflow/core/viewstate"
)

const installedSkillsSnapshotName = "installed_skills"
const toolPresenceSnapshotName = "tool_presence"
const allStarSkillsSnapshotName = "all_star_skills"

func (a *App) ensureViewCache() *viewstate.Manager {
	if a.viewCache != nil {
		return a.viewCache
	}

	root := a.cacheDir
	if strings.TrimSpace(root) == "" {
		root = filepath.Join(config.AppDataDir(), "cache")
	}
	a.viewCache = viewstate.NewManager(filepath.Join(root, "viewstate"))
	return a.viewCache
}

func (a *App) listSkillsUncached() ([]InstalledSkillEntry, error) {
	skills, err := a.storage.ListAll()
	if err != nil {
		return nil, err
	}
	for _, sk := range skills {
		if sk.Category == "" {
			sk.Category = defaultCategoryName
		}
	}
	installedIndex := skill.BuildInstalledIndex(skills)
	return a.buildInstalledSkillEntries(skills, a.buildToolPresenceIndex(installedIndex)), nil
}

func (a *App) installedSkillsFingerprint() (string, error) {
	cfg, err := a.config.Load()
	if err != nil {
		return "", err
	}

	metaDir := filepath.Join(filepath.Dir(filepath.Clean(cfg.SkillsStorageDir)), "meta")
	metaFingerprint, err := directorySummaryFingerprint(metaDir)
	if err != nil {
		return "", err
	}

	toolsConfig, err := json.Marshal(cfg.Tools)
	if err != nil {
		return "", err
	}

	parts := []string{
		strconv.Itoa(cfg.RepoScanMaxDepth),
		string(toolsConfig),
		metaFingerprint,
	}
	for _, tool := range cfg.Tools {
		pushDirFingerprint, err := directorySummaryFingerprint(tool.PushDir)
		if err != nil {
			return "", err
		}
		parts = append(parts, tool.Name, pushDirFingerprint)
	}

	return viewstate.HashFingerprint(parts...), nil
}

func (a *App) toolPresenceConfigFingerprint(cfg config.AppConfig) (string, error) {
	toolsConfig, err := json.Marshal(struct {
		RepoScanMaxDepth int                 `json:"repoScanMaxDepth"`
		Tools            []config.ToolConfig `json:"tools"`
	}{
		RepoScanMaxDepth: cfg.RepoScanMaxDepth,
		Tools:            cfg.Tools,
	})
	if err != nil {
		return "", err
	}
	return viewstate.HashFingerprint(string(toolsConfig)), nil
}

func (a *App) allStarSkillsFingerprint() (string, error) {
	repos, err := a.starStorage.Load()
	if err != nil {
		return "", err
	}

	installedFingerprint, err := a.installedSkillsFingerprint()
	if err != nil {
		return "", err
	}

	repoData, err := json.Marshal(struct {
		RepoScanMaxDepth int                   `json:"repoScanMaxDepth"`
		Repos            []coregit.StarredRepo `json:"repos"`
	}{
		RepoScanMaxDepth: a.repoScanMaxDepth(),
		Repos:            repos,
	})
	if err != nil {
		return "", err
	}

	return viewstate.HashFingerprint(installedFingerprint, string(repoData)), nil
}

func (a *App) buildToolPresenceSnapshot(cfg config.AppConfig, idx *skill.InstalledIndex) (viewstate.ToolPresenceSnapshot, error) {
	configFingerprint, err := a.toolPresenceConfigFingerprint(cfg)
	if err != nil {
		return viewstate.ToolPresenceSnapshot{}, err
	}

	var previous viewstate.ToolPresenceSnapshot
	_, _ = a.ensureViewCache().Load(toolPresenceSnapshotName, configFingerprint, &previous)

	inputs := make([]viewstate.ToolPresenceInput, 0, len(cfg.Tools))
	toolsByName := make(map[string]config.ToolConfig, len(cfg.Tools))
	for _, tool := range cfg.Tools {
		if strings.TrimSpace(tool.PushDir) == "" {
			continue
		}
		fingerprint, err := directorySummaryFingerprint(tool.PushDir)
		if err != nil {
			return viewstate.ToolPresenceSnapshot{}, err
		}
		inputs = append(inputs, viewstate.ToolPresenceInput{
			Name:        tool.Name,
			Fingerprint: fingerprint,
		})
		toolsByName[tool.Name] = tool
	}

	next, err := viewstate.RebuildToolPresence(previous, inputs, func(toolName string) ([]string, error) {
		tool, ok := toolsByName[toolName]
		if !ok {
			return nil, nil
		}
		if _, err := os.Stat(tool.PushDir); err != nil {
			if os.IsNotExist(err) {
				return nil, nil
			}
			return nil, err
		}

		pushed, err := pullToolSkills(a.ctx, getAdapter(tool), tool.PushDir, a.repoScanMaxDepth())
		if err != nil {
			return nil, err
		}

		var keys []string
		for _, candidate := range pushed {
			keys = append(keys, toolPresenceKeys(candidate, idx)...)
		}
		return compactKeys(keys...), nil
	})
	if err != nil {
		return viewstate.ToolPresenceSnapshot{}, err
	}

	_ = a.ensureViewCache().Save(toolPresenceSnapshotName, configFingerprint, next)
	return next, nil
}

func directorySummaryFingerprint(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return viewstate.HashFingerprint("empty"), nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return viewstate.HashFingerprint("missing", filepath.Clean(path)), nil
		}
		return "", err
	}

	latestModTime := int64(0)
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
		info, err := entry.Info()
		if err != nil {
			return "", err
		}
		if modTime := info.ModTime().UnixNano(); modTime > latestModTime {
			latestModTime = modTime
		}
	}
	sort.Strings(names)

	return viewstate.HashFingerprint(
		filepath.Clean(path),
		strconv.Itoa(len(entries)),
		strconv.FormatInt(latestModTime, 10),
		strings.Join(names, "\x00"),
	), nil
}
