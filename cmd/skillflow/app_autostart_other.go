//go:build !windows && !darwin

package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

const xdgAutostartTemplate = `[Desktop Entry]
Type=Application
Name={{.DisplayName}}
Exec={{.ExecLine}}
X-GNOME-Autostart-enabled=true
`

type xdgLaunchAtLoginController struct {
	exePath      string
	autostartDir string
}

func newLaunchAtLoginController(exePath string) (launchAtLoginController, error) {
	configHome := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME"))
	if configHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		configHome = filepath.Join(homeDir, ".config")
	}
	return &xdgLaunchAtLoginController{
		exePath:      exePath,
		autostartDir: filepath.Join(configHome, "autostart"),
	}, nil
}

func (c *xdgLaunchAtLoginController) path() string {
	return filepath.Join(c.autostartDir, launchAtLoginAppName+".desktop")
}

func (c *xdgLaunchAtLoginController) IsEnabled() bool {
	_, err := os.Stat(c.path())
	return err == nil
}

func (c *xdgLaunchAtLoginController) Enable() error {
	if err := os.MkdirAll(c.autostartDir, 0o755); err != nil {
		return err
	}
	file, err := os.Create(c.path())
	if err != nil {
		return err
	}
	defer file.Close()

	tpl := template.Must(template.New("xdg-autostart").Parse(xdgAutostartTemplate))
	return tpl.Execute(file, struct {
		DisplayName string
		ExecLine    string
	}{
		DisplayName: launchAtLoginAppName,
		ExecLine:    quoteDesktopExec([]string{c.exePath}),
	})
}

func (c *xdgLaunchAtLoginController) Disable() error {
	return os.Remove(c.path())
}

func quoteDesktopExec(args []string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = strconv.Quote(arg)
	}
	return strings.Join(quoted, " ")
}
