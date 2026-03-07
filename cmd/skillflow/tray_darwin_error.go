//go:build darwin

package main

import "errors"

var errDarwinTraySetup = errors.New("darwin tray setup failed")
