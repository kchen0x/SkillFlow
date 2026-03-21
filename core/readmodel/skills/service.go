package skills

import (
	"context"
	"fmt"
	"strings"

	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
	"github.com/shinerio/skillflow/core/readmodel/viewstate"
	skillquery "github.com/shinerio/skillflow/core/skillcatalog/app/query"
	skilldomain "github.com/shinerio/skillflow/core/skillcatalog/domain"
	sourcedomain "github.com/shinerio/skillflow/core/skillsource/domain"
)

type InstalledSkillProvider interface {
	ListAll() ([]*skilldomain.InstalledSkill, error)
}

type SourceCandidateProvider interface {
	ListAllSourceCandidates(maxDepth int) ([]sourcedomain.SourceSkillCandidate, error)
	ListSourceCandidatesByRepo(repoURL string, maxDepth int) ([]sourcedomain.SourceSkillCandidate, error)
}

type PresenceProvider interface {
	Resolve(ctx context.Context, idx *skillquery.InstalledIndex, maxDepth int, profiles []agentdomain.AgentProfile) (*agentdomain.AgentPresenceIndex, error)
}

type Service struct {
	installed InstalledSkillProvider
	sources   SourceCandidateProvider
	presence  PresenceProvider
	cache     *viewstate.Manager
}

func NewService(installed InstalledSkillProvider, sources SourceCandidateProvider, presence PresenceProvider, cache *viewstate.Manager) *Service {
	return &Service{
		installed: installed,
		sources:   sources,
		presence:  presence,
		cache:     cache,
	}
}

func (s *Service) ListInstalledSkills(ctx context.Context, input InstalledSkillsInput) ([]InstalledSkillEntry, error) {
	if cached, hit, err := s.loadInstalledSkillsSnapshot(input.SnapshotFingerprint); err != nil {
		return nil, err
	} else if hit {
		return cached, nil
	}

	installed, idx, err := s.loadInstalledIndex()
	if err != nil {
		return nil, err
	}
	presence, err := s.resolvePresence(ctx, idx, input.RepoScanMaxDepth, input.AgentProfiles)
	if err != nil {
		return nil, err
	}
	entries := buildInstalledSkillEntries(installed, presence, input.DefaultCategory)
	if entries == nil {
		entries = []InstalledSkillEntry{}
	}

	if err := s.saveSnapshot(InstalledSkillsSnapshotName, input.SnapshotFingerprint, entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func (s *Service) ListAllStarSkills(ctx context.Context, input StarSkillsInput) ([]StarSkillEntry, error) {
	if cached, hit, err := s.loadAllStarSkillsSnapshot(input.SnapshotFingerprint); err != nil {
		return nil, err
	} else if hit {
		return cached, nil
	}

	if s.sources == nil {
		return nil, fmt.Errorf("skills readmodel sources provider is not configured")
	}
	candidates, err := s.sources.ListAllSourceCandidates(input.RepoScanMaxDepth)
	if err != nil {
		return nil, err
	}
	_, idx, err := s.loadInstalledIndex()
	if err != nil {
		return nil, err
	}
	presence, err := s.resolvePresence(ctx, idx, input.RepoScanMaxDepth, input.AgentProfiles)
	if err != nil {
		return nil, err
	}
	resolved := resolveSourceCandidates(candidates, idx, presence)
	if resolved == nil {
		resolved = []StarSkillEntry{}
	}

	if err := s.saveSnapshot(AllStarSkillsSnapshotName, input.SnapshotFingerprint, resolved); err != nil {
		return nil, err
	}
	return resolved, nil
}

func (s *Service) ListRepoStarSkills(ctx context.Context, repoURL string, input StarSkillsInput) ([]StarSkillEntry, error) {
	if s.sources == nil {
		return nil, fmt.Errorf("skills readmodel sources provider is not configured")
	}
	candidates, err := s.sources.ListSourceCandidatesByRepo(repoURL, input.RepoScanMaxDepth)
	if err != nil {
		return nil, err
	}
	_, idx, err := s.loadInstalledIndex()
	if err != nil {
		return nil, err
	}
	presence, err := s.resolvePresence(ctx, idx, input.RepoScanMaxDepth, input.AgentProfiles)
	if err != nil {
		return nil, err
	}

	resolved := resolveSourceCandidates(candidates, idx, presence)
	if resolved == nil {
		return []StarSkillEntry{}, nil
	}
	return resolved, nil
}

func (s *Service) loadInstalledIndex() ([]*skilldomain.InstalledSkill, *skillquery.InstalledIndex, error) {
	if s.installed == nil {
		return nil, nil, fmt.Errorf("skills readmodel installed provider is not configured")
	}
	installed, err := s.installed.ListAll()
	if err != nil {
		return nil, nil, err
	}
	return installed, skillquery.BuildInstalledIndex(installed), nil
}

func (s *Service) resolvePresence(ctx context.Context, idx *skillquery.InstalledIndex, maxDepth int, profiles []agentdomain.AgentProfile) (*agentdomain.AgentPresenceIndex, error) {
	if s.presence == nil {
		return agentdomain.NewAgentPresenceIndex(), nil
	}
	presence, err := s.presence.Resolve(ctx, idx, maxDepth, profiles)
	if err != nil {
		return nil, err
	}
	if presence == nil {
		return agentdomain.NewAgentPresenceIndex(), nil
	}
	return presence, nil
}

func (s *Service) loadInstalledSkillsSnapshot(fingerprint string) ([]InstalledSkillEntry, bool, error) {
	if s.cache == nil || strings.TrimSpace(fingerprint) == "" {
		return nil, false, nil
	}
	var cached []InstalledSkillEntry
	state, err := s.cache.Load(InstalledSkillsSnapshotName, fingerprint, &cached)
	if err != nil {
		return nil, false, err
	}
	if state == viewstate.StateHit {
		return cached, true, nil
	}
	return nil, false, nil
}

func (s *Service) loadAllStarSkillsSnapshot(fingerprint string) ([]StarSkillEntry, bool, error) {
	if s.cache == nil || strings.TrimSpace(fingerprint) == "" {
		return nil, false, nil
	}
	var cached []StarSkillEntry
	state, err := s.cache.Load(AllStarSkillsSnapshotName, fingerprint, &cached)
	if err != nil {
		return nil, false, err
	}
	if state == viewstate.StateHit {
		return cached, true, nil
	}
	return nil, false, nil
}

func (s *Service) saveSnapshot(name, fingerprint string, payload any) error {
	if s.cache == nil || strings.TrimSpace(fingerprint) == "" {
		return nil
	}
	return s.cache.Save(name, fingerprint, payload)
}
