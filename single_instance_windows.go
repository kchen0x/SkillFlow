//go:build windows

package main

import (
	"os"
	"syscall"
	"unsafe"
)

const swRestore = 9

// ensureSingleInstance uses a named mutex to guarantee only one instance runs.
// If another instance is already running, it brings that window to the foreground
// and exits the current process.
func ensureSingleInstance() {
	name, _ := syscall.UTF16PtrFromString("Local\\SkillFlowSingleInstance")
	handle, _, err := procCreateMutexW.Call(0, 0, uintptr(unsafe.Pointer(name)))
	if handle == 0 {
		// Could not create mutex — allow startup to proceed.
		return
	}
	if err == syscall.ERROR_ALREADY_EXISTS {
		// Another instance is running: find its window and bring it to front.
		title, _ := syscall.UTF16PtrFromString("SkillFlow")
		hwnd, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(title)))
		if hwnd != 0 {
			iconic, _, _ := procIsIconic.Call(hwnd)
			if iconic != 0 {
				procShowWindow.Call(hwnd, swRestore)
			}
			procSetForegroundWindow.Call(hwnd)
		}
		os.Exit(0)
	}
}
