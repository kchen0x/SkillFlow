package orchestration

import (
	"context"
	"path/filepath"

	agentapp "github.com/shinerio/skillflow/core/agentintegration/app"
	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
	"github.com/shinerio/skillflow/core/shared/logicalkey"
	skillquery "github.com/shinerio/skillflow/core/skillcatalog/app/query"
	skilldomain "github.com/shinerio/skillflow/core/skillcatalog/domain"
)

type ImportLocalCommand struct {
	SourceDir          string
	Category           string
	AgentProfiles      []agentdomain.AgentProfile
	AutoPushAgentNames []string
	TriggerAutoBackup  bool
}

type ImportLocalResult struct {
	Skill    *skilldomain.InstalledSkill
	AutoPush AutoPushReport
	Backup   BackupReport
}

func (s *Service) ImportLocalSkill(ctx context.Context, cmd ImportLocalCommand) (ImportLocalResult, error) {
	if err := requireDependency(s.skillCatalog, "skill catalog service"); err != nil {
		return ImportLocalResult{}, err
	}
	source := "local.import"
	category := s.normalizeCategory(cmd.Category)
	s.logger.Infof("orchestration import local started: source=%s dir=%s category=%s", source, cmd.SourceDir, category)
	skill, err := s.skillCatalog.Import(cmd.SourceDir, category, skilldomain.SourceManual, "", "")
	if err != nil {
		s.logger.Errorf("orchestration import local failed: source=%s dir=%s category=%s err=%v", source, cmd.SourceDir, category, err)
		return ImportLocalResult{}, err
	}
	result := ImportLocalResult{Skill: skill}
	result.AutoPush = s.autoPushSkills(ctx, source, cmd.AgentProfiles, cmd.AutoPushAgentNames, []*skilldomain.InstalledSkill{skill}, false)
	result.Backup = s.triggerAutoBackup(ctx, source, cmd.TriggerAutoBackup)
	s.logger.Infof("orchestration import local completed: source=%s skillID=%s name=%s", source, skill.ID, skill.Name)
	return result, nil
}

type ImportRepoSourceSkillsCommand struct {
	SkillPaths         []string
	RepoRootDir        string
	CanonicalRepoURL   string
	Category           string
	AgentProfiles      []agentdomain.AgentProfile
	AutoPushAgentNames []string
	TriggerAutoBackup  bool
}

type ImportRepoSourceSkillsResult struct {
	Imported []*skilldomain.InstalledSkill
	AutoPush AutoPushReport
	Backup   BackupReport
}

func (s *Service) ImportRepoSourceSkills(ctx context.Context, cmd ImportRepoSourceSkillsCommand) (ImportRepoSourceSkillsResult, error) {
	if err := requireDependency(s.skillCatalog, "skill catalog service"); err != nil {
		return ImportRepoSourceSkillsResult{}, err
	}
	source := "starred.import"
	category := s.normalizeCategory(cmd.Category)
	s.logger.Infof("orchestration import repo source skills started: source=%s repo=%s selected=%d category=%s", source, cmd.CanonicalRepoURL, len(cmd.SkillPaths), category)

	installed, err := s.skillCatalog.ListAll()
	if err != nil {
		s.logger.Errorf("orchestration import repo source skills failed: source=%s repo=%s list installed failed: %v", source, cmd.CanonicalRepoURL, err)
		return ImportRepoSourceSkillsResult{}, err
	}
	idx := skillquery.BuildInstalledIndex(installed)

	imported := make([]*skilldomain.InstalledSkill, 0, len(cmd.SkillPaths))
	for _, skillPath := range cmd.SkillPaths {
		subPath, err := filepath.Rel(cmd.RepoRootDir, skillPath)
		if err != nil {
			return ImportRepoSourceSkillsResult{}, err
		}
		subPath = filepath.ToSlash(subPath)
		logicalKey := logicalkey.GitFromRepoURLOrEmpty(cmd.CanonicalRepoURL, subPath)
		if idx.IsInstalled(filepath.Base(skillPath), logicalKey) {
			continue
		}

		skill, importErr := s.skillCatalog.Import(skillPath, category, skilldomain.SourceGitHub, cmd.CanonicalRepoURL, subPath)
		if importErr != nil {
			if s.isSkillExistsError(importErr) {
				continue
			}
			s.logger.Errorf("orchestration import repo source skills failed: source=%s repo=%s path=%s err=%v", source, cmd.CanonicalRepoURL, skillPath, importErr)
			return ImportRepoSourceSkillsResult{}, importErr
		}

		if s.resolveRepoSubPathSHA != nil {
			sha, err := s.resolveRepoSubPathSHA(coalesceContext(ctx), cmd.RepoRootDir, subPath)
			if err == nil {
				skill.SourceSHA = sha
				_ = s.skillCatalog.UpdateMeta(skill)
			}
		}
		imported = append(imported, skill)
	}

	result := ImportRepoSourceSkillsResult{Imported: imported}
	result.AutoPush = s.autoPushSkills(ctx, source, cmd.AgentProfiles, cmd.AutoPushAgentNames, imported, false)
	result.Backup = s.triggerAutoBackup(ctx, source, cmd.TriggerAutoBackup)
	s.logger.Infof("orchestration import repo source skills completed: source=%s repo=%s imported=%d", source, cmd.CanonicalRepoURL, len(imported))
	return result, nil
}

type PushInstalledSkillsCommand struct {
	SkillIDs      []string
	AgentNames    []string
	AgentProfiles []agentdomain.AgentProfile
	Force         bool
}

type PushInstalledSkillsResult struct {
	SelectedSkills []*skilldomain.InstalledSkill
	Conflicts      []agentdomain.PushConflict
}

func (s *Service) PushInstalledSkills(ctx context.Context, cmd PushInstalledSkillsCommand) (PushInstalledSkillsResult, error) {
	if err := requireDependency(s.skillCatalog, "skill catalog service"); err != nil {
		return PushInstalledSkillsResult{}, err
	}
	if err := requireDependency(s.agentIntegration, "agent integration service"); err != nil {
		return PushInstalledSkillsResult{}, err
	}
	s.logger.Infof("orchestration push skills started: force=%t skillCount=%d agentCount=%d", cmd.Force, len(cmd.SkillIDs), len(cmd.AgentNames))
	allSkills, err := s.skillCatalog.ListAll()
	if err != nil {
		s.logger.Errorf("orchestration push skills failed: force=%t list installed failed: %v", cmd.Force, err)
		return PushInstalledSkillsResult{}, err
	}
	idSet := map[string]struct{}{}
	for _, skillID := range cmd.SkillIDs {
		idSet[skillID] = struct{}{}
	}
	selected := make([]*skilldomain.InstalledSkill, 0, len(cmd.SkillIDs))
	for _, skill := range allSkills {
		if skill == nil {
			continue
		}
		if _, ok := idSet[skill.ID]; ok {
			selected = append(selected, skill)
		}
	}
	conflicts, err := s.agentIntegration.PushSkills(coalesceContext(ctx), cmd.AgentProfiles, cmd.AgentNames, selected, cmd.Force)
	if err != nil {
		s.logger.Errorf("orchestration push skills failed: force=%t selected=%d err=%v", cmd.Force, len(selected), err)
		return PushInstalledSkillsResult{}, err
	}
	s.logger.Infof("orchestration push skills completed: force=%t selected=%d conflicts=%d", cmd.Force, len(selected), len(conflicts))
	return PushInstalledSkillsResult{
		SelectedSkills: selected,
		Conflicts:      conflicts,
	}, nil
}

type PullFromAgentCommand struct {
	AgentName          string
	SkillPaths         []string
	Category           string
	AgentProfiles      []agentdomain.AgentProfile
	RepoScanMaxDepth   int
	Force              bool
	AutoPushAgentNames []string
	TriggerAutoBackup  bool
}

type PullFromAgentResult struct {
	AgentFound bool
	Conflicts  []string
	Imported   []*skilldomain.InstalledSkill
	AutoPush   AutoPushReport
	Backup     BackupReport
}

func (s *Service) PullFromAgent(ctx context.Context, cmd PullFromAgentCommand) (PullFromAgentResult, error) {
	if err := requireDependency(s.skillCatalog, "skill catalog service"); err != nil {
		return PullFromAgentResult{}, err
	}
	if err := requireDependency(s.agentIntegration, "agent integration service"); err != nil {
		return PullFromAgentResult{}, err
	}
	category := s.normalizeCategory(cmd.Category)
	source := "agent.pull"
	if cmd.Force {
		source = "agent.pull.force"
	}
	s.logger.Infof("orchestration pull from agent started: source=%s agent=%s selectedPaths=%d category=%s", source, cmd.AgentName, len(cmd.SkillPaths), category)
	profile, found := findAgentProfile(cmd.AgentProfiles, cmd.AgentName)
	if !found {
		s.logger.Infof("orchestration pull from agent skipped: source=%s agent=%s reason=agent-not-found", source, cmd.AgentName)
		return PullFromAgentResult{}, nil
	}
	installed, err := s.skillCatalog.ListAll()
	if err != nil {
		s.logger.Errorf("orchestration pull from agent failed: source=%s agent=%s list installed failed: %v", source, cmd.AgentName, err)
		return PullFromAgentResult{}, err
	}
	idx := skillquery.BuildInstalledIndex(installed)
	presence, err := s.agentIntegration.BuildPresenceIndex(coalesceContext(ctx), cmd.AgentProfiles, idx, normalizeScanDepth(cmd.RepoScanMaxDepth))
	if err != nil {
		s.logger.Errorf("orchestration pull from agent failed: source=%s agent=%s build presence failed: %v", source, cmd.AgentName, err)
		return PullFromAgentResult{}, err
	}
	candidates, err := s.agentIntegration.ScanAgentSkills(coalesceContext(ctx), profile, idx, presence, normalizeScanDepth(cmd.RepoScanMaxDepth))
	if err != nil {
		s.logger.Errorf("orchestration pull from agent failed: source=%s agent=%s scan failed: %v", source, cmd.AgentName, err)
		return PullFromAgentResult{}, err
	}
	selected := agentapp.SelectAgentSkillCandidates(candidates, cmd.SkillPaths)
	imported := make([]*skilldomain.InstalledSkill, 0, len(selected))
	conflicts := make([]string, 0)
	for _, candidate := range selected {
		if !cmd.Force && candidate.Imported {
			conflicts = append(conflicts, candidate.Path)
			continue
		}
		if cmd.Force {
			if err := s.deleteConflictingInstalledSkills(candidate); err != nil {
				s.logger.Errorf("orchestration pull from agent failed: source=%s agent=%s delete existing failed: skill=%s err=%v", source, cmd.AgentName, candidate.Name, err)
				return PullFromAgentResult{}, err
			}
		}
		skill, importErr := s.skillCatalog.Import(candidate.Path, category, skilldomain.SourceManual, "", "")
		if importErr != nil {
			if !cmd.Force && s.isSkillExistsError(importErr) {
				conflicts = append(conflicts, candidate.Path)
				continue
			}
			s.logger.Errorf("orchestration pull from agent failed: source=%s agent=%s import path=%s err=%v", source, cmd.AgentName, candidate.Path, importErr)
			return PullFromAgentResult{}, importErr
		}
		imported = append(imported, skill)
	}

	result := PullFromAgentResult{
		AgentFound: true,
		Conflicts:  conflicts,
		Imported:   imported,
	}
	result.AutoPush = s.autoPushSkills(ctx, source, cmd.AgentProfiles, cmd.AutoPushAgentNames, imported, cmd.Force)
	result.Backup = s.triggerAutoBackup(ctx, source, cmd.TriggerAutoBackup)
	s.logger.Infof("orchestration pull from agent completed: source=%s agent=%s imported=%d conflicts=%d", source, cmd.AgentName, len(imported), len(conflicts))
	return result, nil
}

func (s *Service) deleteConflictingInstalledSkills(candidate agentdomain.AgentSkillCandidate) error {
	existing, err := s.skillCatalog.ListAll()
	if err != nil {
		return err
	}
	for _, installed := range existing {
		if installed == nil {
			continue
		}
		if matchesCandidate(installed, candidate) {
			// Keep force semantics aligned with current shell behavior: best-effort delete.
			_ = s.skillCatalog.Delete(installed.ID)
		}
	}
	return nil
}

func matchesCandidate(installed *skilldomain.InstalledSkill, candidate agentdomain.AgentSkillCandidate) bool {
	if installed == nil {
		return false
	}
	if candidate.LogicalKey != "" {
		logicalKey, err := skilldomain.LogicalKey(installed)
		return err == nil && logicalKey == candidate.LogicalKey
	}
	return installed.Name == candidate.Name
}
