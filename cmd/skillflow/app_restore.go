package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	platformgit "github.com/shinerio/skillflow/core/platform/git"
	skilldomain "github.com/shinerio/skillflow/core/skillcatalog/domain"
)

type cloudRestoreState struct {
	installedSkills map[string]cloudRestoreSkillSnapshot
	starredRepoURLs map[string]struct{}
}

type cloudRestoreSkillSnapshot struct {
	SourceSHA string
	UpdatedAt time.Time
}

func (a *App) captureCloudRestoreState() cloudRestoreState {
	state := cloudRestoreState{
		installedSkills: map[string]cloudRestoreSkillSnapshot{},
		starredRepoURLs: map[string]struct{}{},
	}

	if a.storage != nil {
		if skills, err := a.storage.ListAll(); err == nil {
			for _, sk := range skills {
				if key := cloudRestoreSkillKey(sk); key != "" {
					state.installedSkills[key] = cloudRestoreSkillSnapshot{
						SourceSHA: strings.TrimSpace(sk.SourceSHA),
						UpdatedAt: sk.UpdatedAt,
					}
				}
			}
		}
	}

	if a.starStorage != nil {
		if repos, err := a.starStorage.Load(); err == nil {
			for _, repo := range repos {
				if key := cloudRestoreRepoKey(repo.URL); key != "" {
					state.starredRepoURLs[key] = struct{}{}
				}
			}
		}
	}

	return state
}

func (a *App) handleRestoredCloudState(before cloudRestoreState, source string) error {
	a.logInfof("restore compensation started: source=%s", source)
	a.reloadStateFromDisk()

	restoredSkills, err := a.restoredOrUpdatedSkillsSince(before)
	if err != nil {
		a.logErrorf("restore compensation failed: source=%s load restored skills failed: %v", source, err)
		return err
	}
	if len(restoredSkills) > 0 {
		a.autoPushSkillsToConfiguredAgents(source, restoredSkills, true)
	}

	clonedRepos, failedRepos, err := a.cloneNewlyRestoredStarredRepos(before, source)
	if err != nil {
		a.logErrorf("restore compensation failed: source=%s clone restored starred repos failed: %v", source, err)
		return err
	}

	a.logInfof("restore compensation completed: source=%s restoredSkills=%d clonedRepos=%d failedRepos=%d", source, len(restoredSkills), clonedRepos, failedRepos)
	return nil
}

func (a *App) restoredOrUpdatedSkillsSince(before cloudRestoreState) ([]*skilldomain.InstalledSkill, error) {
	if a.storage == nil {
		return nil, nil
	}

	skills, err := a.storage.ListAll()
	if err != nil {
		return nil, err
	}
	restored := make([]*skilldomain.InstalledSkill, 0, len(skills))
	for _, sk := range skills {
		key := cloudRestoreSkillKey(sk)
		if key == "" {
			continue
		}
		snapshot, existed := before.installedSkills[key]
		if existed && snapshot.SourceSHA == strings.TrimSpace(sk.SourceSHA) && snapshot.UpdatedAt.Equal(sk.UpdatedAt) {
			continue
		}
		restored = append(restored, sk)
	}
	return restored, nil
}

func (a *App) cloneNewlyRestoredStarredRepos(before cloudRestoreState, source string) (int, int, error) {
	if a.starStorage == nil {
		return 0, 0, nil
	}

	repos, err := a.starStorage.Load()
	if err != nil {
		return 0, 0, err
	}

	cloned := 0
	failed := 0
	changed := false
	for i := range repos {
		repoKey := cloudRestoreRepoKey(repos[i].URL)
		if repoKey == "" {
			continue
		}
		if _, existed := before.starredRepoURLs[repoKey]; existed {
			continue
		}
		if hasClonedRepo(repos[i].LocalDir) {
			continue
		}

		a.logInfof("restore starred repo clone started: source=%s repo=%s localDir=%s", source, repos[i].URL, repos[i].LocalDir)
		if err := platformgit.CloneOrUpdate(a.cloneContext(), repos[i].URL, repos[i].LocalDir, a.gitProxyURL()); err != nil {
			failed++
			changed = true
			repos[i].SyncError = err.Error()
			a.logErrorf("restore starred repo clone failed: source=%s repo=%s localDir=%s err=%v", source, repos[i].URL, repos[i].LocalDir, err)
			continue
		}

		cloned++
		changed = true
		repos[i].SyncError = ""
		repos[i].LastSync = time.Now()
		a.logInfof("restore starred repo clone completed: source=%s repo=%s localDir=%s", source, repos[i].URL, repos[i].LocalDir)
	}

	if changed {
		if err := a.starStorage.Save(repos); err != nil {
			return cloned, failed, err
		}
	}
	return cloned, failed, nil
}

func cloudRestoreSkillKey(sk *skilldomain.InstalledSkill) string {
	if sk == nil {
		return ""
	}
	if logicalKey, err := skilldomain.LogicalKey(sk); err == nil && strings.TrimSpace(logicalKey) != "" {
		return "logical:" + logicalKey
	}
	if strings.TrimSpace(sk.ID) == "" {
		return ""
	}
	return "instance:" + strings.TrimSpace(sk.ID)
}

func cloudRestoreRepoKey(repoURL string) string {
	if normalized, err := platformgit.CanonicalRepoURL(repoURL); err == nil && strings.TrimSpace(normalized) != "" {
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

func (a *App) cloneContext() context.Context {
	if a.ctx != nil {
		return a.ctx
	}
	return context.Background()
}
