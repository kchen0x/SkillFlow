package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/emersion/go-autostart"
)

type launchAtLoginController interface {
	IsEnabled() bool
	Enable() error
	Disable() error
}

func (a *App) autostartController() (launchAtLoginController, error) {
	if a.autostartFactory != nil {
		return a.autostartFactory()
	}
	return a.autostartApp()
}

func (a *App) autostartApp() (*autostart.App, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("resolve executable path failed: %w", err)
	}
	return &autostart.App{
		Name:        "SkillFlow",
		DisplayName: "SkillFlow",
		Exec:        []string{exePath},
	}, nil
}

func (a *App) syncLaunchAtLogin(enabled bool) error {
	a.logInfof("launch-at-login update started: desired=%t", enabled)
	app, err := a.autostartController()
	if err != nil {
		a.logErrorf("launch-at-login update failed: desired=%t err=%v", enabled, err)
		return err
	}
	if enabled {
		if err := app.Enable(); err != nil {
			a.logErrorf("launch-at-login update failed: desired=%t action=enable err=%v", enabled, err)
			return err
		}
		a.logInfof("launch-at-login update completed: desired=%t action=enable", enabled)
		return nil
	}
	if !app.IsEnabled() {
		a.logInfof("launch-at-login update completed: desired=%t action=noop", enabled)
		return nil
	}
	if err := app.Disable(); err != nil && !errors.Is(err, os.ErrNotExist) {
		a.logErrorf("launch-at-login update failed: desired=%t action=disable err=%v", enabled, err)
		return err
	} else {
		a.logInfof("launch-at-login update completed: desired=%t action=disable", enabled)
	}
	return nil
}
