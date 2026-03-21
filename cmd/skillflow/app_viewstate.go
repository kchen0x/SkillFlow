package main

import (
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"

	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
	readmodelskills "github.com/shinerio/skillflow/core/readmodel/skills"
	"github.com/shinerio/skillflow/core/readmodel/viewstate"
	skillquery "github.com/shinerio/skillflow/core/skillcatalog/app/query"
	sourcedomain "github.com/shinerio/skillflow/core/skillsource/domain"
)

const (
	installedSkillsSnapshotName = readmodelskills.InstalledSkillsSnapshotName
	agentPresenceSnapshotName   = readmodelskills.AgentPresenceSnapshotName
	allStarSkillsSnapshotName   = readmodelskills.AllStarSkillsSnapshotName
)

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

func (a *App) installedSkillsFingerprint() (string, error) {
	dataDir := a.dataDir()
	metaDir := filepath.Join(dataDir, "meta")
	metaFingerprint, err := directorySummaryFingerprint(metaDir)
	if err != nil {
		return "", err
	}
	metaLocalDir := filepath.Join(dataDir, "meta_local")
	metaLocalFingerprint, err := directorySummaryFingerprint(metaLocalDir)
	if err != nil {
		return "", err
	}

	cfg, err := a.config.Load()
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
		RepoScanMaxDepth int                     `json:"repoScanMaxDepth"`
		Repos            []sourcedomain.StarRepo `json:"repos"`
	}{
		RepoScanMaxDepth: a.repoScanMaxDepth(),
		Repos:            repos,
	})
	if err != nil {
		return "", err
	}

	return viewstate.HashFingerprint(installedFingerprint, string(repoData)), nil
}

func (a *App) buildAgentPresenceIndex(idx *skillquery.InstalledIndex) *agentdomain.AgentPresenceIndex {
	presence, _ := measureOperation(a, "build_agent_presence_index", func() (*agentdomain.AgentPresenceIndex, error) {
		cfg, err := a.config.Load()
		if err != nil {
			a.logErrorf("build agent presence index failed: load config err=%v", err)
			return agentdomain.NewAgentPresenceIndex(), nil
		}
		presence, err := a.newPresenceResolver().Resolve(a.ctx, idx, a.repoScanMaxDepth(), cfg.Agents)
		if err != nil {
			a.logErrorf("build agent presence index failed: resolve presence err=%v", err)
			return agentdomain.NewAgentPresenceIndex(), nil
		}
		return presence, nil
	})
	return presence
}

func directorySummaryFingerprint(path string) (string, error) {
	return readmodelskills.DirectorySummaryFingerprint(path)
}
