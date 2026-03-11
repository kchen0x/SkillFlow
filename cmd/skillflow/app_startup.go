package main

import "time"

type startupBackgroundTask struct {
	name  string
	delay time.Duration
	run   func()
}

func scheduleStartupBackgroundTasks(tasks []startupBackgroundTask, schedule func(startupBackgroundTask)) {
	for _, task := range tasks {
		schedule(task)
	}
}

func (a *App) startupBackgroundTaskPlan() []startupBackgroundTask {
	return []startupBackgroundTask{
		{name: "git.pull", delay: 750 * time.Millisecond, run: a.gitPullOnStartup},
		{name: "skills.check_updates", delay: 3 * time.Second, run: a.checkUpdatesOnStartup},
		{name: "starred.refresh", delay: 5250 * time.Millisecond, run: a.updateStarredReposOnStartup},
		{name: "app.check_update", delay: 8 * time.Second, run: a.checkAppUpdateOnStartup},
	}
}
