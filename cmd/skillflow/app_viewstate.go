package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
	"github.com/shinerio/skillflow/core/config"
	coregit "github.com/shinerio/skillflow/core/git"
	skillquery "github.com/shinerio/skillflow/core/skillcatalog/app/query"
	"github.com/shinerio/skillflow/core/viewstate"
)

const installedSkillsSnapshotName = "installed_skills"
const agentPresenceSnapshotName = "agent_presence"
const allStarSkillsSnapshotName = "all_star_skills"

func (a *App) ensureViewCache() *viewstate.Manager {
	if a.viewCache != nil {
		return a.viewCache
	}

	root := a.cacheDir
	if strings.TrimSpace(root) == "" {
		root = filepath.Join(appDataDirFunc(), "cache")
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
	installedIndex := skillquery.BuildInstalledIndex(skills)
	return a.buildInstalledSkillEntries(skills, a.buildAgentPresenceIndex(installedIndex)), nil
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
	metaLocalDir := filepath.Join(filepath.Dir(filepath.Clean(cfg.SkillsStorageDir)), "meta_local")
	metaLocalFingerprint, err := directorySummaryFingerprint(metaLocalDir)
	if err != nil {
		return "", err
	}

	agentsConfig, err := json.Marshal(cfg.Agents)
	if err != nil {
		return "", err
	}

	parts := []string{
		strconv.Itoa(cfg.RepoScanMaxDepth),
		string(agentsConfig),
		metaFingerprint,
		metaLocalFingerprint,
	}
	for _, agent := range cfg.Agents {
		pushDirFingerprint, err := directorySummaryFingerprint(agent.PushDir)
		if err != nil {
			return "", err
		}
		parts = append(parts, agent.Name, pushDirFingerprint)
	}

	return viewstate.HashFingerprint(parts...), nil
}

func (a *App) agentPresenceConfigFingerprint(cfg config.AppConfig) (string, error) {
	agentsConfig, err := json.Marshal(struct {
		RepoScanMaxDepth int                  `json:"repoScanMaxDepth"`
		Agents           []config.AgentConfig `json:"agents"`
	}{
		RepoScanMaxDepth: cfg.RepoScanMaxDepth,
		Agents:           cfg.Agents,
	})
	if err != nil {
		return "", err
	}
	return viewstate.HashFingerprint(string(agentsConfig)), nil
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

func (a *App) buildAgentPresenceSnapshot(cfg config.AppConfig, idx *skillquery.InstalledIndex) (viewstate.AgentPresenceSnapshot, error) {
	configFingerprint, err := a.agentPresenceConfigFingerprint(cfg)
	if err != nil {
		return viewstate.AgentPresenceSnapshot{}, err
	}

	var previous viewstate.AgentPresenceSnapshot
	_, _ = a.ensureViewCache().Load(agentPresenceSnapshotName, configFingerprint, &previous)

	inputs := make([]viewstate.AgentPresenceInput, 0, len(cfg.Agents))
	agentsByName := make(map[string]config.AgentConfig, len(cfg.Agents))
	for _, agent := range cfg.Agents {
		if strings.TrimSpace(agent.PushDir) == "" {
			continue
		}
		fingerprint, err := directorySummaryFingerprint(agent.PushDir)
		if err != nil {
			return viewstate.AgentPresenceSnapshot{}, err
		}
		inputs = append(inputs, viewstate.AgentPresenceInput{
			Name:        agent.Name,
			Fingerprint: fingerprint,
		})
		agentsByName[agent.Name] = agent
	}

	next, err := viewstate.RebuildAgentPresence(previous, inputs, func(agentName string) ([]string, error) {
		agent, ok := agentsByName[agentName]
		if !ok {
			return nil, nil
		}
		presence, err := newAgentIntegrationService().BuildPresenceIndex(a.ctx, []agentdomain.AgentProfile{agentProfile(agent)}, idx, a.repoScanMaxDepth())
		if err != nil {
			return nil, err
		}
		return presence.KeysForAgent(agent.Name), nil
	})
	if err != nil {
		return viewstate.AgentPresenceSnapshot{}, err
	}

	_ = a.ensureViewCache().Save(agentPresenceSnapshotName, configFingerprint, next)
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
