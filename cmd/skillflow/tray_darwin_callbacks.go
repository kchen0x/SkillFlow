//go:build darwin

package main

/*
 */
import "C"

//export skillflowTrayOnShow
func skillflowTrayOnShow() {
	go withDarwinTrayApp(func(app *App) {
		app.showMainWindow()
	})
}

//export skillflowTrayOnHide
func skillflowTrayOnHide() {
	go withDarwinTrayApp(func(app *App) {
		app.hideMainWindow()
	})
}

//export skillflowTrayOnQuit
func skillflowTrayOnQuit() {
	go withDarwinTrayApp(func(app *App) {
		app.quitApp()
	})
}

//export skillflowTrayOnApplicationWillHide
func skillflowTrayOnApplicationWillHide() {
	go withDarwinTrayApp(func(app *App) {
		app.logInfof("main window hide started, mode=menu_bar")
	})
}

//export skillflowTrayOnApplicationDidHide
func skillflowTrayOnApplicationDidHide() {
	go withDarwinTrayApp(func(app *App) {
		applyDarwinAccessoryPolicy()
		app.publishWindowVisibilityChanged(false)
		app.logInfof("main window hide completed, mode=menu_bar")
	})
}
