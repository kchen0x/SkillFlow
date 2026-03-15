package main

import (
	"context"
	goruntime "runtime"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	setupTrayForUIFn    = setupTray
	teardownTrayForUIFn = teardownTray
)

func uiProcessOwnsTrayLifecycle() bool {
	return goruntime.GOOS == "darwin" && !helperBootstrapEnabled()
}

func (a *App) setupTrayForUI(ctx context.Context) {
	if !uiProcessOwnsTrayLifecycle() {
		return
	}
	if err := setupTrayForUIFn(a); err != nil {
		if ctx != nil {
			runtime.LogWarningf(ctx, "tray init failed: %v", err)
			return
		}
		a.logErrorf("tray init failed: %v", err)
	}
}

func (a *App) teardownTrayForUI() {
	if !uiProcessOwnsTrayLifecycle() {
		return
	}
	teardownTrayForUIFn()
}
