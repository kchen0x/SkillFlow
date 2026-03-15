//go:build darwin && !bindings

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunEntryStartsUIProcessDirectlyOnDarwin(t *testing.T) {
	prevRunUIProcessFn := runUIProcessFn
	prevBootstrapHelperProcessFn := bootstrapHelperProcessFn
	t.Cleanup(func() {
		runUIProcessFn = prevRunUIProcessFn
		bootstrapHelperProcessFn = prevBootstrapHelperProcessFn
	})

	runUICalls := 0
	helperCalls := 0
	runUIProcessFn = func() error {
		runUICalls++
		return nil
	}
	bootstrapHelperProcessFn = func(_ []string) error {
		helperCalls++
		return nil
	}

	exitCode := runEntry([]string{"SkillFlow"})

	assert.Equal(t, 0, exitCode)
	assert.Equal(t, 1, runUICalls)
	assert.Equal(t, 0, helperCalls)
}

func TestBuildUIOptionsHideWindowOnCloseWhenUIOwnsTrayLifecycle(t *testing.T) {
	opts := buildUIOptions(NewApp())
	require.NotNil(t, opts)
	assert.True(t, opts.HideWindowOnClose)
}
