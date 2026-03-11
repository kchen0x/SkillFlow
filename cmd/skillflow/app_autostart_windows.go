//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type windowsLaunchAtLoginController struct {
	exePath    string
	startupDir string
}

func newLaunchAtLoginController(exePath string) (launchAtLoginController, error) {
	startupDir, err := windowsStartupDir()
	if err != nil {
		return nil, err
	}
	return &windowsLaunchAtLoginController{
		exePath:    exePath,
		startupDir: startupDir,
	}, nil
}

func windowsStartupDir() (string, error) {
	if appData := strings.TrimSpace(os.Getenv("APPDATA")); appData != "" {
		return filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Startup"), nil
	}
	if userProfile := strings.TrimSpace(os.Getenv("USERPROFILE")); userProfile != "" {
		return filepath.Join(userProfile, "AppData", "Roaming", "Microsoft", "Windows", "Start Menu", "Programs", "Startup"), nil
	}
	return "", fmt.Errorf("resolve Windows startup directory failed: APPDATA and USERPROFILE are not set")
}

func (c *windowsLaunchAtLoginController) path() string {
	return filepath.Join(c.startupDir, launchAtLoginAppName+".cmd")
}

func (c *windowsLaunchAtLoginController) IsEnabled() bool {
	_, err := os.Stat(c.path())
	return err == nil
}

func (c *windowsLaunchAtLoginController) Enable() error {
	if err := os.MkdirAll(c.startupDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(c.path(), []byte(windowsLaunchAtLoginScript(c.exePath)), 0o644)
}

func (c *windowsLaunchAtLoginController) Disable() error {
	return os.Remove(c.path())
}

func windowsLaunchAtLoginScript(exePath string) string {
	escapedPath := strings.ReplaceAll(exePath, "\"", "\"\"")
	return "@echo off\r\nstart \"\" \"" + escapedPath + "\"\r\n"
}
