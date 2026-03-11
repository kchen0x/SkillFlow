package main

type trayController interface {
	showMainWindow()
	hideMainWindow()
	quitApp()
	logDebugf(format string, args ...any)
	logInfof(format string, args ...any)
	logErrorf(format string, args ...any)
}

type trayVisibilityPublisher interface {
	publishWindowVisibilityChanged(visible bool)
}
