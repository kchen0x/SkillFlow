//go:build !windows

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

type fileHelperInstanceLock struct {
	path string
}

func acquireHelperInstanceLock() (helperInstanceLock, error) {
	lockPath := filepath.Join(runtimeStateDir(), "helper.lock")
	if err := os.MkdirAll(filepath.Dir(lockPath), 0755); err != nil {
		return nil, err
	}

	if err := writeHelperLockFile(lockPath); err == nil {
		return &fileHelperInstanceLock{path: lockPath}, nil
	} else if !errors.Is(err, os.ErrExist) {
		return nil, err
	}

	pid, err := readHelperLockPID(lockPath)
	if err == nil && pid > 0 && processRunning(pid) {
		return nil, errHelperAlreadyRunning
	}

	if removeErr := os.Remove(lockPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
		return nil, removeErr
	}
	if err := writeHelperLockFile(lockPath); err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil, errHelperAlreadyRunning
		}
		return nil, err
	}
	return &fileHelperInstanceLock{path: lockPath}, nil
}

func (l *fileHelperInstanceLock) Release() error {
	if l == nil || strings.TrimSpace(l.path) == "" {
		return nil
	}
	err := os.Remove(l.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func writeHelperLockFile(path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(fmt.Sprintf("%d", os.Getpid())); err != nil {
		return err
	}
	return nil
}

func readHelperLockPID(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, err
	}
	return pid, nil
}

func processRunning(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = proc.Signal(syscall.Signal(0))
	return err == nil
}
