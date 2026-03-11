//go:build bindings

package main

import "testing"

func TestHelperBootstrapDisabledInBindingsMode(t *testing.T) {
	if helperBootstrapEnabled() {
		t.Fatal("expected helper bootstrap to be disabled in bindings mode")
	}
}
