//go:build darwin

package main

/*
 */
import "C"

import "strings"

//export skillflowTrayLog
func skillflowTrayLog(level *C.char, message *C.char) {
	withDarwinTrayApp(func(app *App) {
		msg := C.GoString(message)
		switch strings.ToLower(C.GoString(level)) {
		case "debug":
			app.logDebugf("%s", msg)
		case "error":
			app.logErrorf("%s", msg)
		default:
			app.logInfof("%s", msg)
		}
	})
}
