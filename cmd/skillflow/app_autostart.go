package main

import (
	"fmt"
	"os"

	"github.com/emersion/go-autostart"
)

func (a *App) autostartApp() (autostart.App, error) {
	exePath, err := os.Executable()
	if err != nil {
		return autostart.App{}, fmt.Errorf("resolve executable path failed: %w", err)
	}
	return autostart.App{
		Name:        "SkillFlow",
		DisplayName: "SkillFlow",
		Exec:        []string{exePath},
	}, nil
}

func (a *App) syncLaunchAtLogin(enabled bool) error {
	a.logInfof("launch-at-login update started: enabled=%t", enabled)
	app, err := a.autostartApp()
	if err != nil {
		a.logErrorf("launch-at-login update failed: enabled=%t err=%v", enabled, err)
		return err
	}
	if enabled {
		if err := app.Enable(); err != nil {
			a.logErrorf("launch-at-login update failed: enabled=%t err=%v", enabled, err)
			return err
		}
	} else {
		if err := app.Disable(); err != nil {
			a.logErrorf("launch-at-login update failed: enabled=%t err=%v", enabled, err)
			return err
		}
	}
	a.logInfof("launch-at-login update completed: enabled=%t", enabled)
	return nil
}
