//go:build darwin && !bindings

package main

import "testing"

func TestHelperBootstrapDisabledOnDarwin(t *testing.T) {
	if helperBootstrapEnabled() {
		t.Fatal("expected helper bootstrap to be disabled on darwin to avoid duplicate SkillFlow processes")
	}
}
