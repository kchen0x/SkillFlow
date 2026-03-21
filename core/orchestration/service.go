package orchestration

import (
	"context"
	"fmt"
	"strings"
	"time"

	agentapp "github.com/shinerio/skillflow/core/agentintegration/app"
	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
	skillquery "github.com/shinerio/skillflow/core/skillcatalog/app/query"
	skilldomain "github.com/shinerio/skillflow/core/skillcatalog/domain"
	sourcedomain "github.com/shinerio/skillflow/core/skillsource/domain"
)

const (
	defaultCategoryName      = "Default"
	defaultRepoScanMaxDepth  = 5
	defaultRestoreSourceName = "restore.compensation"
)

type SkillCatalogService interface {
	Import(srcDir, category string, source skilldomain.SourceType, sourceURL, sourceSubPath string) (*skilldomain.InstalledSkill, error)
	Get(id string) (*skilldomain.InstalledSkill, error)
	ListAll() ([]*skilldomain.InstalledSkill, error)
	Delete(id string) error
	OverwriteFromDir(id, srcDir string) error
	UpdateMeta(sk *skilldomain.InstalledSkill) error
}

type AgentIntegrationService interface {
	PushSkills(ctx context.Context, profiles []agentdomain.AgentProfile, agentNames []string, skills []*skilldomain.InstalledSkill, force bool) ([]agentdomain.PushConflict, error)
	BuildPresenceIndex(ctx context.Context, profiles []agentdomain.AgentProfile, idx *skillquery.InstalledIndex, maxDepth int) (*agentdomain.AgentPresenceIndex, error)
	ScanAgentSkills(ctx context.Context, profile agentdomain.AgentProfile, idx *skillquery.InstalledIndex, presence *agentdomain.AgentPresenceIndex, maxDepth int) ([]agentdomain.AgentSkillCandidate, error)
	RefreshPushedCopies(ctx context.Context, profiles []agentdomain.AgentProfile, skill *skilldomain.InstalledSkill) error
}

type SkillSourceResolver interface {
	ResolveCachedSource(ctx context.Context, skill *skilldomain.InstalledSkill) (sourceDir, latestSHA string, err error)
}

type StarRepoStore interface {
	Load() ([]sourcedomain.StarRepo, error)
	Save([]sourcedomain.StarRepo) error
}

type RepoCloner interface {
	CloneOrUpdate(ctx context.Context, repoURL, dir, proxyURL string) error
}

type AutoBackupScheduler interface {
	ScheduleAutoBackup(ctx context.Context, source string) error
}

type Logger interface {
	Infof(format string, args ...any)
	Errorf(format string, args ...any)
	Debugf(format string, args ...any)
}

type SkillExistsErrorMatcher func(err error) bool
type RepoSubPathSHAResolver func(ctx context.Context, repoDir, subPath string) (string, error)

type Dependencies struct {
	SkillCatalog          SkillCatalogService
	AgentIntegration      AgentIntegrationService
	SkillSource           SkillSourceResolver
	StarRepoStore         StarRepoStore
	RepoCloner            RepoCloner
	AutoBackup            AutoBackupScheduler
	IsSkillExistsError    SkillExistsErrorMatcher
	ResolveRepoSubPathSHA RepoSubPathSHAResolver
	Logger                Logger
	Now                   func() time.Time
	DefaultCategory       string
}

type Service struct {
	skillCatalog          SkillCatalogService
	agentIntegration      AgentIntegrationService
	skillSource           SkillSourceResolver
	starRepoStore         StarRepoStore
	repoCloner            RepoCloner
	autoBackup            AutoBackupScheduler
	isSkillExistsError    SkillExistsErrorMatcher
	resolveRepoSubPathSHA RepoSubPathSHAResolver
	logger                Logger
	now                   func() time.Time
	defaultCategory       string
}

func NewService(deps Dependencies) *Service {
	now := deps.Now
	if now == nil {
		now = time.Now
	}
	category := strings.TrimSpace(deps.DefaultCategory)
	if category == "" {
		category = defaultCategoryName
	}
	matcher := deps.IsSkillExistsError
	if matcher == nil {
		matcher = func(error) bool { return false }
	}
	lg := deps.Logger
	if lg == nil {
		lg = noopLogger{}
	}
	return &Service{
		skillCatalog:          deps.SkillCatalog,
		agentIntegration:      deps.AgentIntegration,
		skillSource:           deps.SkillSource,
		starRepoStore:         deps.StarRepoStore,
		repoCloner:            deps.RepoCloner,
		autoBackup:            deps.AutoBackup,
		isSkillExistsError:    matcher,
		resolveRepoSubPathSHA: deps.ResolveRepoSubPathSHA,
		logger:                lg,
		now:                   now,
		defaultCategory:       category,
	}
}

type noopLogger struct{}

func (noopLogger) Infof(string, ...any)  {}
func (noopLogger) Errorf(string, ...any) {}
func (noopLogger) Debugf(string, ...any) {}

type AutoPushReport struct {
	Attempted bool
	Source    string
	Force     bool
	Agents    []string
	Conflicts []agentdomain.PushConflict
	Err       error
}

type BackupReport struct {
	Triggered bool
	Source    string
	Err       error
}

func (s *Service) autoPushSkills(ctx context.Context, source string, profiles []agentdomain.AgentProfile, autoPushAgentNames []string, skills []*skilldomain.InstalledSkill, force bool) AutoPushReport {
	if len(skills) == 0 {
		return AutoPushReport{Source: source, Force: force}
	}
	targets := autoPushAgentTargets(profiles, autoPushAgentNames)
	if len(targets) == 0 {
		s.logger.Debugf("orchestration auto-push skipped: source=%s reason=no-target-agents", source)
		return AutoPushReport{Source: source, Force: force}
	}
	report := AutoPushReport{
		Attempted: true,
		Source:    source,
		Force:     force,
		Agents:    agentProfileNames(targets),
	}
	if s.agentIntegration == nil {
		report.Err = fmt.Errorf("agent integration service is not configured")
		s.logger.Errorf("orchestration auto-push failed: source=%s reason=missing-agent-service", source)
		return report
	}
	conflicts, err := s.agentIntegration.PushSkills(coalesceContext(ctx), targets, report.Agents, skills, force)
	if err != nil {
		report.Err = err
		s.logger.Errorf("orchestration auto-push failed: source=%s agentCount=%d skillCount=%d force=%t err=%v", source, len(report.Agents), len(skills), force, err)
		return report
	}
	report.Conflicts = conflicts
	s.logger.Infof("orchestration auto-push completed: source=%s agentCount=%d skillCount=%d conflicts=%d force=%t", source, len(report.Agents), len(skills), len(conflicts), force)
	return report
}

func autoPushAgentTargets(profiles []agentdomain.AgentProfile, autoPushAgents []string) []agentdomain.AgentProfile {
	if len(autoPushAgents) == 0 {
		return nil
	}
	selected := make(map[string]struct{}, len(autoPushAgents))
	for _, name := range autoPushAgents {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		selected[name] = struct{}{}
	}
	targets := make([]agentdomain.AgentProfile, 0, len(selected))
	for _, profile := range profiles {
		if !profile.Enabled {
			continue
		}
		if _, ok := selected[profile.Name]; ok {
			targets = append(targets, profile)
		}
	}
	return targets
}

func agentProfileNames(profiles []agentdomain.AgentProfile) []string {
	result := make([]string, 0, len(profiles))
	for _, profile := range profiles {
		result = append(result, profile.Name)
	}
	return result
}

func (s *Service) triggerAutoBackup(ctx context.Context, source string, enabled bool) BackupReport {
	if !enabled {
		return BackupReport{Source: source}
	}
	if s.autoBackup == nil {
		s.logger.Debugf("orchestration auto-backup skipped: source=%s reason=missing-scheduler", source)
		return BackupReport{Source: source}
	}
	s.logger.Infof("orchestration auto-backup started: source=%s", source)
	report := BackupReport{
		Triggered: true,
		Source:    source,
	}
	if err := s.autoBackup.ScheduleAutoBackup(coalesceContext(ctx), source); err != nil {
		report.Err = err
		s.logger.Errorf("orchestration auto-backup failed: source=%s err=%v", source, err)
		return report
	}
	s.logger.Infof("orchestration auto-backup completed: source=%s", source)
	return report
}

func (s *Service) normalizeCategory(category string) string {
	trimmed := strings.TrimSpace(category)
	if trimmed == "" || strings.EqualFold(trimmed, s.defaultCategory) {
		return s.defaultCategory
	}
	return trimmed
}

func normalizeScanDepth(depth int) int {
	if depth <= 0 {
		return defaultRepoScanMaxDepth
	}
	return depth
}

func coalesceContext(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	return context.Background()
}

func requireDependency(dep any, name string) error {
	if dep == nil {
		return fmt.Errorf("%s is not configured", name)
	}
	return nil
}

func findAgentProfile(profiles []agentdomain.AgentProfile, agentName string) (agentdomain.AgentProfile, bool) {
	return agentapp.FindProfile(profiles, strings.TrimSpace(agentName))
}
