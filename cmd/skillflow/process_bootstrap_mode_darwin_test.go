//go:build darwin && !bindings

package main

import "testing"

func TestHelperBootstrapEnabledOnDarwin(t *testing.T) {
	if !helperBootstrapEnabled() {
		t.Fatal("expected helper bootstrap to be enabled on darwin so window close can release the UI process")
	}
}
