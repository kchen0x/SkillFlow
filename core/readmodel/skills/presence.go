package skills

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
	"github.com/shinerio/skillflow/core/readmodel/viewstate"
	skillquery "github.com/shinerio/skillflow/core/skillcatalog/app/query"
)

type PresenceBuilder interface {
	BuildPresenceIndex(ctx context.Context, profiles []agentdomain.AgentProfile, idx *skillquery.InstalledIndex, maxDepth int) (*agentdomain.AgentPresenceIndex, error)
}

type PresenceResolver struct {
	cache   *viewstate.Manager
	builder PresenceBuilder
}

func NewPresenceResolver(cache *viewstate.Manager, builder PresenceBuilder) *PresenceResolver {
	return &PresenceResolver{
		cache:   cache,
		builder: builder,
	}
}

func (r *PresenceResolver) Resolve(ctx context.Context, idx *skillquery.InstalledIndex, maxDepth int, profiles []agentdomain.AgentProfile) (*agentdomain.AgentPresenceIndex, error) {
	if idx == nil {
		idx = skillquery.BuildInstalledIndex(nil)
	}

	configFingerprint, err := agentPresenceConfigFingerprint(maxDepth, profiles)
	if err != nil {
		return nil, err
	}

	var previous viewstate.AgentPresenceSnapshot
	if r.cache != nil {
		_, _ = r.cache.Load(AgentPresenceSnapshotName, configFingerprint, &previous)
	}

	agentsByName := make(map[string]agentdomain.AgentProfile, len(profiles))
	inputs := make([]viewstate.AgentPresenceInput, 0, len(profiles))
	for _, profile := range profiles {
		if strings.TrimSpace(profile.PushDir) == "" {
			continue
		}

		fingerprint, err := DirectorySummaryFingerprint(profile.PushDir)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, viewstate.AgentPresenceInput{
			Name:        profile.Name,
			Fingerprint: fingerprint,
		})
		agentsByName[profile.Name] = profile
	}

	next, err := viewstate.RebuildAgentPresence(previous, inputs, func(agentName string) ([]string, error) {
		if r.builder == nil {
			return nil, nil
		}
		profile, ok := agentsByName[agentName]
		if !ok {
			return nil, nil
		}
		presence, err := r.builder.BuildPresenceIndex(ctx, []agentdomain.AgentProfile{profile}, idx, maxDepth)
		if err != nil {
			return nil, err
		}
		return presence.KeysForAgent(agentName), nil
	})
	if err != nil {
		return nil, err
	}

	if r.cache != nil {
		_ = r.cache.Save(AgentPresenceSnapshotName, configFingerprint, next)
	}
	return presenceIndexFromSnapshot(next), nil
}

func DirectorySummaryFingerprint(path string) (string, error) {
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

func agentPresenceConfigFingerprint(maxDepth int, profiles []agentdomain.AgentProfile) (string, error) {
	encoded, err := json.Marshal(struct {
		RepoScanMaxDepth int                        `json:"repoScanMaxDepth"`
		Agents           []agentdomain.AgentProfile `json:"agents"`
	}{
		RepoScanMaxDepth: maxDepth,
		Agents:           profiles,
	})
	if err != nil {
		return "", err
	}
	return viewstate.HashFingerprint(string(encoded)), nil
}

func presenceIndexFromSnapshot(snapshot viewstate.AgentPresenceSnapshot) *agentdomain.AgentPresenceIndex {
	presence := agentdomain.NewAgentPresenceIndex()
	for key, agentNames := range snapshot.AgentsByKey {
		for _, agentName := range agentNames {
			presence.Add(agentName, key)
		}
	}
	return presence
}
