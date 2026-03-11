//go:build darwin

package main

/*
 */
import "C"

//export skillflowTrayOnShow
func skillflowTrayOnShow() {
	go withDarwinTrayController(func(controller trayController) {
		controller.showMainWindow()
	})
}

//export skillflowTrayOnHide
func skillflowTrayOnHide() {
	go withDarwinTrayController(func(controller trayController) {
		controller.hideMainWindow()
	})
}

//export skillflowTrayOnQuit
func skillflowTrayOnQuit() {
	go withDarwinTrayController(func(controller trayController) {
		controller.quitApp()
	})
}

//export skillflowTrayOnApplicationWillHide
func skillflowTrayOnApplicationWillHide() {
	go withDarwinTrayController(func(controller trayController) {
		controller.logInfof("main window hide started, mode=menu_bar")
	})
}

//export skillflowTrayOnApplicationDidHide
func skillflowTrayOnApplicationDidHide() {
	go withDarwinTrayController(func(controller trayController) {
		applyDarwinAccessoryPolicy()
		if publisher, ok := controller.(trayVisibilityPublisher); ok {
			publisher.publishWindowVisibilityChanged(false)
		}
		controller.logInfof("main window hide completed, mode=menu_bar")
	})
}
