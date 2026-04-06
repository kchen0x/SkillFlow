//go:build darwin && !bindings

package main

import (
	"context"
	"testing"

	"github.com/shinerio/skillflow/core/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBeforeCloseRequestsRuntimeQuitForHelperManagedMacUI(t *testing.T) {
	prevActiveProcessRole := activeProcessRole
	prevRuntimeQuitFn := runtimeQuitFn
	prevWindowGetSizeFn := windowGetSizeFn
	t.Cleanup(func() {
		activeProcessRole = prevActiveProcessRole
		runtimeQuitFn = prevRuntimeQuitFn
		windowGetSizeFn = prevWindowGetSizeFn
	})

	activeProcessRole = processRoleUI

	quitCalls := 0
	runtimeQuitFn = func(ctx context.Context) {
		quitCalls++
		require.NotNil(t, ctx)
	}
	windowGetSizeFn = func(context.Context) (int, int) {
		return 1440, 920
	}

	app := NewApp()
	app.config = config.NewService(t.TempDir())
	app.initialWindowState = config.WindowState{Width: 1360, Height: 860}

	preventClose := app.beforeClose(context.Background())

	assert.True(t, preventClose)
	assert.Equal(t, 1, quitCalls)
	assert.True(t, app.uiQuitInProgress())
	assert.True(t, app.windowVisibilityInit)
	assert.False(t, app.windowVisible)

	state, ok := app.config.LoadWindowState()
	require.True(t, ok)
	assert.Equal(t, config.NormalizeWindowState(config.WindowState{Width: 1440, Height: 920}), state)
}

func TestBeforeCloseAllowsQuitWhenHelperManagedMacUIQuitAlreadyInProgress(t *testing.T) {
	prevActiveProcessRole := activeProcessRole
	prevRuntimeQuitFn := runtimeQuitFn
	prevWindowGetSizeFn := windowGetSizeFn
	t.Cleanup(func() {
		activeProcessRole = prevActiveProcessRole
		runtimeQuitFn = prevRuntimeQuitFn
		windowGetSizeFn = prevWindowGetSizeFn
	})

	activeProcessRole = processRoleUI
	runtimeQuitFn = func(context.Context) {
		t.Fatal("runtime quit should not be requested twice")
	}
	windowGetSizeFn = func(context.Context) (int, int) {
		t.Fatal("window size should not be read again while quit is already in progress")
		return 0, 0
	}

	app := NewApp()
	app.uiQuitRequested = true

	preventClose := app.beforeClose(context.Background())

	assert.False(t, preventClose)
}

func TestQuitAppMarksHelperManagedMacUIQuitInProgress(t *testing.T) {
	prevActiveProcessRole := activeProcessRole
	prevRuntimeQuitFn := runtimeQuitFn
	t.Cleanup(func() {
		activeProcessRole = prevActiveProcessRole
		runtimeQuitFn = prevRuntimeQuitFn
	})

	activeProcessRole = processRoleUI

	quitCalls := 0
	runtimeQuitFn = func(ctx context.Context) {
		quitCalls++
		require.NotNil(t, ctx)
	}

	app := NewApp()
	app.ctx = context.Background()

	app.quitApp()

	assert.Equal(t, 1, quitCalls)
	assert.True(t, app.uiQuitInProgress())
}
