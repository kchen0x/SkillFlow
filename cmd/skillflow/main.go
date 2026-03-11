package main

import (
	"embed"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	os.Exit(runEntry(os.Args))
}

func runEntry(args []string) int {
	if !helperBootstrapEnabled() {
		if err := runUIProcess(); err != nil {
			println("Error:", err.Error())
			return 1
		}
		return 0
	}

	role, filteredArgs := determineProcessRole(args)
	activeProcessRole = role
	if len(filteredArgs) > 0 {
		os.Args = filteredArgs
	}
	if role == processRoleUI {
		if err := runUIProcess(); err != nil {
			println("Error:", err.Error())
			return 1
		}
		return 0
	}
	uiArgs := []string(nil)
	if len(filteredArgs) > 1 {
		uiArgs = filteredArgs[1:]
	}
	if err := bootstrapHelperProcess(uiArgs); err != nil {
		println("Error:", err.Error())
		return 1
	}
	return 0
}

func runUIProcess() error {
	app := NewApp()

	return wails.Run(&options.App{
		Title:             "SkillFlow",
		Width:             1360,
		Height:            860,
		MinWidth:          960,
		MinHeight:         680,
		HideWindowOnClose: false,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnBeforeClose:    app.beforeClose,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})
}
