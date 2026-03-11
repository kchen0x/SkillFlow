package main

import "errors"

var errHelperAlreadyRunning = errors.New("helper already running")

type helperInstanceLock interface {
	Release() error
}
