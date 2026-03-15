//go:build darwin && !bindings

package main

func helperBootstrapEnabled() bool {
	return false
}
