//go:build darwin && !bindings

package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupTrayForUIDoesNothingWhenHelperOwnsLifecycle(t *testing.T) {
	prevSetupTrayForUIFn := setupTrayForUIFn
	t.Cleanup(func() {
		setupTrayForUIFn = prevSetupTrayForUIFn
	})

	calls := 0
	setupTrayForUIFn = func(_ trayController) error {
		calls++
		return nil
	}

	app := NewApp()
	app.setupTrayForUI(context.Background())

	assert.Equal(t, 0, calls)
}

func TestTeardownTrayForUIDoesNothingWhenHelperOwnsLifecycle(t *testing.T) {
	prevTeardownTrayForUIFn := teardownTrayForUIFn
	t.Cleanup(func() {
		teardownTrayForUIFn = prevTeardownTrayForUIFn
	})

	calls := 0
	teardownTrayForUIFn = func() {
		calls++
	}

	app := NewApp()
	app.teardownTrayForUI()

	assert.Equal(t, 0, calls)
}
