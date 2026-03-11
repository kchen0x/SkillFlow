package main

import "time"

func measureOperation[T any](app *App, name string, op func() (T, error)) (T, error) {
	start := time.Now()
	value, err := op()

	if app != nil {
		status := "completed"
		if err != nil {
			status = "failed"
		}
		app.logDebugf("performance operation %s: name=%s duration_ms=%d", status, name, time.Since(start).Milliseconds())
	}

	return value, err
}
