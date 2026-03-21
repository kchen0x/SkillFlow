package main

import (
	"context"
	"errors"

	"github.com/shinerio/skillflow/core/orchestration"
	platformgit "github.com/shinerio/skillflow/core/platform/git"
	readmodelskills "github.com/shinerio/skillflow/core/readmodel/skills"
	skillquery "github.com/shinerio/skillflow/core/skillcatalog/app/query"
	skilldomain "github.com/shinerio/skillflow/core/skillcatalog/domain"
	skillrepo "github.com/shinerio/skillflow/core/skillcatalog/infra/repository"
)

type appOrchestrationLogger struct {
	app *App
}

func (l appOrchestrationLogger) Infof(format string, args ...any) {
	l.app.logInfof(format, args...)
}

func (l appOrchestrationLogger) Errorf(format string, args ...any) {
	l.app.logErrorf(format, args...)
}

func (l appOrchestrationLogger) Debugf(format string, args ...any) {
	l.app.logDebugf(format, args...)
}

type appSkillSourceResolver struct {
	app *App
}

func (r appSkillSourceResolver) ResolveCachedSource(_ context.Context, skill *skilldomain.InstalledSkill) (string, string, error) {
	return r.app.cachedSkillSourceDir(skill)
}

type appAutoBackupScheduler struct {
	app *App
}

func (s appAutoBackupScheduler) ScheduleAutoBackup(_ context.Context, _ string) error {
	s.app.scheduleAutoBackup()
	return nil
}

type appRepoCloner struct{}

func (appRepoCloner) CloneOrUpdate(ctx context.Context, repoURL, dir, proxyURL string) error {
	return cloneOrUpdateRepo(ctx, repoURL, dir, proxyURL)
}

func (a *App) newPresenceResolver() *readmodelskills.PresenceResolver {
	return readmodelskills.NewPresenceResolver(a.ensureViewCache(), newAgentIntegrationService())
}

func (a *App) newSkillsReadmodelService() *readmodelskills.Service {
	return readmodelskills.NewService(
		a.storage,
		a.newSkillsourceService(),
		a.newPresenceResolver(),
		a.ensureViewCache(),
	)
}

func (a *App) newOrchestrationService() *orchestration.Service {
	return orchestration.NewService(orchestration.Dependencies{
		SkillCatalog:          a.storage,
		AgentIntegration:      newAgentIntegrationService(),
		SkillSource:           appSkillSourceResolver{app: a},
		StarRepoStore:         a.starStorage,
		RepoCloner:            appRepoCloner{},
		AutoBackup:            appAutoBackupScheduler{app: a},
		IsSkillExistsError:    func(err error) bool { return errors.Is(err, skillrepo.ErrSkillExists) },
		ResolveRepoSubPathSHA: platformgit.GetSubPathSHA,
		Logger:                appOrchestrationLogger{app: a},
		DefaultCategory:       defaultCategoryName,
	})
}

func (a *App) installedIndex() ([]*skilldomain.InstalledSkill, *skillquery.InstalledIndex, error) {
	installed, err := a.storage.ListAll()
	if err != nil {
		return nil, nil, err
	}
	return installed, skillquery.BuildInstalledIndex(installed), nil
}

func (a *App) cloneContext() context.Context {
	if a.ctx != nil {
		return a.ctx
	}
	return context.Background()
}
