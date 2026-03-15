//go:build darwin

package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowMainWindowNativeRestoresRegularWindowLifecycle(t *testing.T) {
	restore := stubDarwinWindowRuntime()
	t.Cleanup(restore)

	calls := make([]string, 0, 5)
	darwinEnsureStatusItemFn = func() {
		calls = append(calls, "ensure_status_item")
	}
	darwinApplyRegularPolicyFn = func() {
		calls = append(calls, "apply_regular_policy")
	}
	darwinRuntimeShowFn = func(context.Context) {
		calls = append(calls, "show")
	}
	darwinRuntimeWindowShowFn = func(context.Context) {
		calls = append(calls, "window_show")
	}
	darwinRuntimeWindowUnminimiseFn = func(context.Context) {
		calls = append(calls, "window_unminimise")
	}

	require.NoError(t, showMainWindowNative(context.Background()))
	assert.Equal(t, []string{
		"ensure_status_item",
		"apply_regular_policy",
		"show",
		"window_show",
		"window_unminimise",
	}, calls)
}

func TestHideMainWindowNativeKeepsStatusItemAliveBeforeHide(t *testing.T) {
	restore := stubDarwinWindowRuntime()
	t.Cleanup(restore)

	calls := make([]string, 0, 2)
	darwinEnsureStatusItemFn = func() {
		calls = append(calls, "ensure_status_item")
	}
	darwinRuntimeHideFn = func(context.Context) {
		calls = append(calls, "hide")
	}

	require.NoError(t, hideMainWindowNative(context.Background()))
	assert.Equal(t, []string{
		"ensure_status_item",
		"hide",
	}, calls)
}

func stubDarwinWindowRuntime() func() {
	prevEnsureStatusItemFn := darwinEnsureStatusItemFn
	prevApplyRegularPolicyFn := darwinApplyRegularPolicyFn
	prevRuntimeShowFn := darwinRuntimeShowFn
	prevRuntimeWindowShowFn := darwinRuntimeWindowShowFn
	prevRuntimeWindowUnminimiseFn := darwinRuntimeWindowUnminimiseFn
	prevRuntimeHideFn := darwinRuntimeHideFn

	return func() {
		darwinEnsureStatusItemFn = prevEnsureStatusItemFn
		darwinApplyRegularPolicyFn = prevApplyRegularPolicyFn
		darwinRuntimeShowFn = prevRuntimeShowFn
		darwinRuntimeWindowShowFn = prevRuntimeWindowShowFn
		darwinRuntimeWindowUnminimiseFn = prevRuntimeWindowUnminimiseFn
		darwinRuntimeHideFn = prevRuntimeHideFn
	}
}
