package main

import (
	"time"

	"github.com/shinerio/skillflow/core/config"
	"github.com/shinerio/skillflow/core/platform/appdata"
	daemonruntime "github.com/shinerio/skillflow/core/platform/daemon"
)

type startupBackgroundTask = daemonruntime.StartupTask

func scheduleStartupBackgroundTasks(tasks []startupBackgroundTask, schedule func(startupBackgroundTask)) {
	for _, task := range tasks {
		schedule(task)
	}
}

func (a *App) startupBackgroundTaskPlan() []startupBackgroundTask {
	return []startupBackgroundTask{
		{Name: "git.pull", Delay: 750 * time.Millisecond, Run: a.gitPullOnStartup},
		{Name: "starred.refresh", Delay: 3 * time.Second, Run: a.updateStarredReposOnStartup},
		{Name: "skills.check_updates", Delay: 5250 * time.Millisecond, Run: a.checkUpdatesOnStartup},
		{Name: "app.check_update", Delay: 8 * time.Second, Run: a.checkAppUpdateOnStartup},
	}
}

func (a *App) startupUILightweight(dataDir string) {
	a.config = config.NewService(dataDir)
	a.initLogger(config.DefaultLogLevel)

	cfg, err := a.config.Load()
	if err != nil {
		a.logErrorf("ui shell startup config load failed: %v", err)
		a.rebuildPathBoundServices(appdata.RepoCacheDir(dataDir))
		return
	}

	a.setLoggerLevel(cfg.LogLevel)
	a.rebuildPathBoundServices(cfg.RepoCacheDir)
	a.logInfof("ui shell startup completed: dataDir=%s", dataDir)
}
