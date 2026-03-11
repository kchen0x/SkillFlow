//go:build darwin

package main

import (
	"os"
	"path/filepath"
	"text/template"
)

const macLaunchAgentTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Label</key>
    <string>{{.Name}}</string>
    <key>ProgramArguments</key>
    <array>
      {{range .Exec -}}
      <string>{{.}}</string>
      {{end}}
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>AbandonProcessGroup</key>
    <true/>
  </dict>
</plist>`

type darwinLaunchAtLoginController struct {
	exePath   string
	launchDir string
}

func newLaunchAtLoginController(exePath string) (launchAtLoginController, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return &darwinLaunchAtLoginController{
		exePath:   exePath,
		launchDir: filepath.Join(homeDir, "Library", "LaunchAgents"),
	}, nil
}

func (c *darwinLaunchAtLoginController) path() string {
	return filepath.Join(c.launchDir, launchAtLoginAppName+".plist")
}

func (c *darwinLaunchAtLoginController) IsEnabled() bool {
	_, err := os.Stat(c.path())
	return err == nil
}

func (c *darwinLaunchAtLoginController) Enable() error {
	if err := os.MkdirAll(c.launchDir, 0o755); err != nil {
		return err
	}
	file, err := os.Create(c.path())
	if err != nil {
		return err
	}
	defer file.Close()

	tpl := template.Must(template.New("launch-agent").Parse(macLaunchAgentTemplate))
	return tpl.Execute(file, struct {
		Name string
		Exec []string
	}{
		Name: launchAtLoginAppName,
		Exec: []string{c.exePath},
	})
}

func (c *darwinLaunchAtLoginController) Disable() error {
	return os.Remove(c.path())
}
