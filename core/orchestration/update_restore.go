package orchestration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
	"github.com/shinerio/skillflow/core/platform/git"
	skilldomain "github.com/shinerio/skillflow/core/skillcatalog/domain"
)

type UpdateInstalledSkillCommand struct {
	SkillID            string
	AgentProfiles      []agentdomain.AgentProfile
	AutoPushAgentNames []string
	TriggerAutoBackup  bool
}

type UpdateInstalledSkillResult struct {
	Skill    *skilldomain.InstalledSkill
	AutoPush AutoPushReport
	Backup   BackupReport
}

func (s *Service) UpdateInstalledSkill(ctx context.Context, cmd UpdateInstalledSkillCommand) (UpdateInstalledSkillResult, error) {
	if err := requireDependency(s.skillCatalog, "skill catalog service"); err != nil {
		return UpdateInstalledSkillResult{}, err
	}
	if err := requireDependency(s.skillSource, "skill source resolver"); err != nil {
		return UpdateInstalledSkillResult{}, err
	}
	if err := requireDependency(s.agentIntegration, "agent integration service"); err != nil {
		return UpdateInstalledSkillResult{}, err
	}

	source := "skill.update"
	s.logger.Infof("orchestration update installed skill started: source=%s skillID=%s", source, cmd.SkillID)
	skill, err := s.skillCatalog.Get(cmd.SkillID)
	if err != nil {
		s.logger.Errorf("orchestration update installed skill failed: source=%s skillID=%s err=%v", source, cmd.SkillID, err)
		return UpdateInstalledSkillResult{}, err
	}

	sourceDir, latestSHA, err := s.skillSource.ResolveCachedSource(coalesceContext(ctx), skill)
	if err != nil {
		s.logger.Errorf("orchestration update installed skill failed: source=%s skillID=%s resolve cached source failed: %v", source, cmd.SkillID, err)
		return UpdateInstalledSkillResult{}, err
	}
	if err := s.skillCatalog.OverwriteFromDir(cmd.SkillID, sourceDir); err != nil {
		s.logger.Errorf("orchestration update installed skill failed: source=%s skillID=%s overwrite failed: %v", source, cmd.SkillID, err)
		return UpdateInstalledSkillResult{}, err
	}

	skill.SourceSHA = latestSHA
	skill.LatestSHA = ""
	_ = s.skillCatalog.UpdateMeta(skill)

	if err := s.agentIntegration.RefreshPushedCopies(coalesceContext(ctx), cmd.AgentProfiles, skill); err != nil {
		s.logger.Errorf("orchestration update installed skill failed: source=%s skillID=%s refresh pushed copies failed: %v", source, cmd.SkillID, err)
		result := UpdateInstalledSkillResult{Skill: skill}
		result.Backup = s.triggerAutoBackup(ctx, source, cmd.TriggerAutoBackup)
		return result, err
	}

	result := UpdateInstalledSkillResult{Skill: skill}
	result.AutoPush = s.autoPushSkills(ctx, source, cmd.AgentProfiles, cmd.AutoPushAgentNames, []*skilldomain.InstalledSkill{skill}, true)
	result.Backup = s.triggerAutoBackup(ctx, source, cmd.TriggerAutoBackup)
	s.logger.Infof("orchestration update installed skill completed: source=%s skillID=%s name=%s", source, cmd.SkillID, skill.Name)
	return result, nil
}

type RestoreSkillSnapshot struct {
	SourceSHA string
	UpdatedAt time.Time
}

type RestoreState struct {
	InstalledSkills map[string]RestoreSkillSnapshot
	StarredRepoURLs map[string]struct{}
}

type RestoreCompensationCommand struct {
	Before             RestoreState
	Source             string
	AgentProfiles      []agentdomain.AgentProfile
	AutoPushAgentNames []string
	ProxyURL           string
}

type RestoreCompensationResult struct {
	RestoredSkills []*skilldomain.InstalledSkill
	AutoPush       AutoPushReport
	ClonedRepos    int
	FailedRepos    int
}

func (s *Service) CaptureRestoreState(context.Context) (RestoreState, error) {
	state := RestoreState{
		InstalledSkills: map[string]RestoreSkillSnapshot{},
		StarredRepoURLs: map[string]struct{}{},
	}

	if s.skillCatalog != nil {
		skills, err := s.skillCatalog.ListAll()
		if err != nil {
			return RestoreState{}, err
		}
		for _, skill := range skills {
			if key := restoreSkillKey(skill); key != "" {
				state.InstalledSkills[key] = RestoreSkillSnapshot{
					SourceSHA: strings.TrimSpace(skill.SourceSHA),
					UpdatedAt: skill.UpdatedAt,
				}
			}
		}
	}

	if s.starRepoStore != nil {
		repos, err := s.starRepoStore.Load()
		if err != nil {
			return RestoreState{}, err
		}
		for _, repo := range repos {
			if key := restoreRepoKey(repo.URL); key != "" {
				state.StarredRepoURLs[key] = struct{}{}
			}
		}
	}
	return state, nil
}

func (s *Service) CompensateRestore(ctx context.Context, cmd RestoreCompensationCommand) (RestoreCompensationResult, error) {
	if err := requireDependency(s.skillCatalog, "skill catalog service"); err != nil {
		return RestoreCompensationResult{}, err
	}
	source := strings.TrimSpace(cmd.Source)
	if source == "" {
		source = defaultRestoreSourceName
	}

	s.logger.Infof("orchestration restore compensation started: source=%s", source)
	restoredSkills, err := s.restoredOrUpdatedSkillsSince(cmd.Before)
	if err != nil {
		s.logger.Errorf("orchestration restore compensation failed: source=%s load restored skills failed: %v", source, err)
		return RestoreCompensationResult{}, err
	}
	result := RestoreCompensationResult{
		RestoredSkills: restoredSkills,
	}
	result.AutoPush = s.autoPushSkills(ctx, source, cmd.AgentProfiles, cmd.AutoPushAgentNames, restoredSkills, true)

	cloned, failed, err := s.cloneNewlyRestoredStarredRepos(ctx, cmd.Before, cmd.ProxyURL, source)
	if err != nil {
		s.logger.Errorf("orchestration restore compensation failed: source=%s clone restored repos failed: %v", source, err)
		return RestoreCompensationResult{}, err
	}
	result.ClonedRepos = cloned
	result.FailedRepos = failed
	s.logger.Infof("orchestration restore compensation completed: source=%s restoredSkills=%d clonedRepos=%d failedRepos=%d", source, len(restoredSkills), cloned, failed)
	return result, nil
}

func (s *Service) restoredOrUpdatedSkillsSince(before RestoreState) ([]*skilldomain.InstalledSkill, error) {
	skills, err := s.skillCatalog.ListAll()
	if err != nil {
		return nil, err
	}
	restored := make([]*skilldomain.InstalledSkill, 0, len(skills))
	for _, skill := range skills {
		key := restoreSkillKey(skill)
		if key == "" {
			continue
		}
		snapshot, existed := before.InstalledSkills[key]
		if existed && snapshot.SourceSHA == strings.TrimSpace(skill.SourceSHA) && snapshot.UpdatedAt.Equal(skill.UpdatedAt) {
			continue
		}
		restored = append(restored, skill)
	}
	return restored, nil
}

func (s *Service) cloneNewlyRestoredStarredRepos(ctx context.Context, before RestoreState, proxyURL, source string) (int, int, error) {
	if s.starRepoStore == nil {
		return 0, 0, nil
	}
	repos, err := s.starRepoStore.Load()
	if err != nil {
		return 0, 0, err
	}

	cloned := 0
	failed := 0
	changed := false
	for i := range repos {
		repoKey := restoreRepoKey(repos[i].URL)
		if repoKey == "" {
			continue
		}
		if _, existed := before.StarredRepoURLs[repoKey]; existed {
			continue
		}
		if hasClonedRepo(repos[i].LocalDir) {
			continue
		}
		if s.repoCloner == nil {
			failed++
			changed = true
			repos[i].SyncError = "repo cloner is not configured"
			continue
		}
		s.logger.Infof("orchestration restore repo clone started: source=%s repo=%s localDir=%s", source, repos[i].URL, repos[i].LocalDir)
		if err := s.repoCloner.CloneOrUpdate(coalesceContext(ctx), repos[i].URL, repos[i].LocalDir, proxyURL); err != nil {
			failed++
			changed = true
			repos[i].SyncError = err.Error()
			s.logger.Errorf("orchestration restore repo clone failed: source=%s repo=%s localDir=%s err=%v", source, repos[i].URL, repos[i].LocalDir, err)
			continue
		}
		cloned++
		changed = true
		repos[i].SyncError = ""
		repos[i].LastSync = s.now()
		s.logger.Infof("orchestration restore repo clone completed: source=%s repo=%s localDir=%s", source, repos[i].URL, repos[i].LocalDir)
	}

	if changed {
		if err := s.starRepoStore.Save(repos); err != nil {
			return cloned, failed, err
		}
	}
	return cloned, failed, nil
}

func restoreSkillKey(skill *skilldomain.InstalledSkill) string {
	if skill == nil {
		return ""
	}
	if logicalKey, err := skilldomain.LogicalKey(skill); err == nil && strings.TrimSpace(logicalKey) != "" {
		return "logical:" + logicalKey
	}
	id := strings.TrimSpace(skill.ID)
	if id == "" {
		return ""
	}
	return "instance:" + id
}

func restoreRepoKey(repoURL string) string {
	if normalized, err := git.CanonicalRepoURL(repoURL); err == nil && strings.TrimSpace(normalized) != "" {
		return normalized
	}
	return strings.TrimSpace(repoURL)
}

func hasClonedRepo(localDir string) bool {
	if strings.TrimSpace(localDir) == "" {
		return false
	}
	if _, err := os.Stat(filepath.Join(localDir, ".git")); err == nil {
		return true
	}
	return false
}
