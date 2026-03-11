package main

import (
	"fmt"
	"os"
)

func (a *App) startUIControlServer() {
	a.uiControlMu.Lock()
	defer a.uiControlMu.Unlock()

	if a.uiControlServer != nil {
		return
	}

	_ = os.Remove(uiControlPath())
	server, err := startLoopbackControlServer(uiControlPath(), a.handleUIControlCommand)
	if err != nil {
		a.logErrorf("ui control server start failed: %v", err)
		return
	}
	a.uiControlServer = server
	a.logInfof("ui control server started")
}

func (a *App) stopUIControlServer() {
	a.uiControlMu.Lock()
	server := a.uiControlServer
	a.uiControlServer = nil
	a.uiControlMu.Unlock()

	if server == nil {
		return
	}
	if err := server.Close(); err != nil {
		a.logErrorf("ui control server stop failed: %v", err)
		return
	}
	a.logInfof("ui control server stopped")
}

func (a *App) handleUIControlCommand(command string) error {
	switch command {
	case controlCommandShow:
		a.showMainWindow()
		return nil
	case controlCommandHide:
		a.hideMainWindow()
		return nil
	case controlCommandQuit:
		a.quitApp()
		return nil
	default:
		return fmt.Errorf("unsupported ui control command: %s", command)
	}
}
