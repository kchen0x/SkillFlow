//go:build darwin

package main

/*
 */
import "C"

import "strings"

//export skillflowTrayLog
func skillflowTrayLog(level *C.char, message *C.char) {
	withDarwinTrayController(func(controller trayController) {
		msg := C.GoString(message)
		switch strings.ToLower(C.GoString(level)) {
		case "debug":
			controller.logDebugf("%s", msg)
		case "error":
			controller.logErrorf("%s", msg)
		default:
			controller.logInfof("%s", msg)
		}
	})
}
