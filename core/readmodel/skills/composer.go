package skills

import (
	"strings"

	agentapp "github.com/shinerio/skillflow/core/agentintegration/app"
	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
	"github.com/shinerio/skillflow/core/shared/logicalkey"
	skillquery "github.com/shinerio/skillflow/core/skillcatalog/app/query"
	skilldomain "github.com/shinerio/skillflow/core/skillcatalog/domain"
	sourcedomain "github.com/shinerio/skillflow/core/skillsource/domain"
)

func buildInstalledSkillEntries(installed []*skilldomain.InstalledSkill, presence *agentdomain.AgentPresenceIndex, defaultCategory string) []InstalledSkillEntry {
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

		category := sk.Category
		if strings.TrimSpace(category) == "" {
			category = defaultCategory
		}

		entries = append(entries, InstalledSkillEntry{
			ID:           sk.ID,
			Name:         sk.Name,
			Path:         sk.Path,
			Category:     category,
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

func resolveSourceCandidates(candidates []sourcedomain.SourceSkillCandidate, idx *skillquery.InstalledIndex, presence *agentdomain.AgentPresenceIndex) []StarSkillEntry {
	resolved := make([]StarSkillEntry, 0, len(candidates))
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate.LogicalKey) == "" {
			candidate.LogicalKey = logicalkey.Git(candidate.Source, candidate.SubPath)
		}

		contentKey, err := logicalkey.ContentFromDir(candidate.Path)
		if err != nil {
			contentKey = ""
		}
		state := agentapp.ResolveSkillStatus(candidate.Name, candidate.LogicalKey, idx, presence, contentKey)

		resolved = append(resolved, StarSkillEntry{
			Name:         candidate.Name,
			Path:         candidate.Path,
			SubPath:      candidate.SubPath,
			RepoURL:      candidate.RepoURL,
			RepoName:     candidate.RepoName,
			Source:       candidate.Source,
			LogicalKey:   coalesceLogicalKey(candidate.LogicalKey, state.LogicalKey),
			Installed:    state.Installed,
			Imported:     state.Imported,
			Updatable:    state.Updatable,
			Pushed:       state.Pushed,
			PushedAgents: state.PushedAgents,
		})
	}
	return resolved
}

func coalesceLogicalKey(primary, secondary string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return secondary
}
