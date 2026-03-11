//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

var procCloseHandle = kernel32.NewProc("CloseHandle")

type windowsHelperInstanceLock struct {
	handle uintptr
}

func acquireHelperInstanceLock() (helperInstanceLock, error) {
	name, _ := syscall.UTF16PtrFromString("Local\\SkillFlowSingleInstance")
	handle, _, err := procCreateMutexW.Call(0, 0, uintptr(unsafe.Pointer(name)))
	if handle == 0 {
		return nil, err
	}
	if err == syscall.ERROR_ALREADY_EXISTS {
		procCloseHandle.Call(handle)
		return nil, errHelperAlreadyRunning
	}
	return &windowsHelperInstanceLock{handle: handle}, nil
}

func (l *windowsHelperInstanceLock) Release() error {
	if l == nil || l.handle == 0 {
		return nil
	}
	procCloseHandle.Call(l.handle)
	l.handle = 0
	return nil
}
