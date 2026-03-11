package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartupBackgroundTaskPlanUsesStaggeredDelays(t *testing.T) {
	app := NewApp()

	tasks := app.startupBackgroundTaskPlan()

	require.Len(t, tasks, 4)
	assert.Equal(t, "git.pull", tasks[0].name)
	assert.Equal(t, 750*time.Millisecond, tasks[0].delay)
	assert.Equal(t, "skills.check_updates", tasks[1].name)
	assert.Equal(t, 3*time.Second, tasks[1].delay)
	assert.Equal(t, "starred.refresh", tasks[2].name)
	assert.Equal(t, 5250*time.Millisecond, tasks[2].delay)
	assert.Equal(t, "app.check_update", tasks[3].name)
	assert.Equal(t, 8*time.Second, tasks[3].delay)
}

func TestScheduleStartupBackgroundTasksRegistersAllTasks(t *testing.T) {
	scheduled := make([]time.Duration, 0, 2)
	executed := make([]string, 0, 2)

	scheduleStartupBackgroundTasks([]startupBackgroundTask{
		{name: "first", delay: 250 * time.Millisecond, run: func() { executed = append(executed, "first") }},
		{name: "second", delay: time.Second, run: func() { executed = append(executed, "second") }},
	}, func(task startupBackgroundTask) {
		scheduled = append(scheduled, task.delay)
		task.run()
	})

	assert.Equal(t, []time.Duration{250 * time.Millisecond, time.Second}, scheduled)
	assert.Equal(t, []string{"first", "second"}, executed)
}
