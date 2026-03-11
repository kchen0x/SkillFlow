//go:build !darwin && !windows

package main

func setupTray(_ trayController) error {
	return nil
}

func teardownTray() {}
