//go:build !bindings && !darwin

package main

import "testing"

func TestHelperBootstrapEnabledByDefault(t *testing.T) {
	if !helperBootstrapEnabled() {
		t.Fatal("expected helper bootstrap to be enabled outside bindings mode")
	}
}
