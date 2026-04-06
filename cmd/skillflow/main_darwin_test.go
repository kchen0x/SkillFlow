//go:build darwin && !bindings

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunEntryStartsDaemonByDefaultOnDarwin(t *testing.T) {
	prevRunUIProcessFn := runUIProcessFn
	prevBootstrapHelperProcessFn := bootstrapHelperProcessFn
	prevRunDaemonProcessFn := runDaemonProcessFn
	t.Cleanup(func() {
		runUIProcessFn = prevRunUIProcessFn
		bootstrapHelperProcessFn = prevBootstrapHelperProcessFn
		runDaemonProcessFn = prevRunDaemonProcessFn
	})

	runUICalls := 0
	daemonCalls := 0
	runUIProcessFn = func() error {
		runUICalls++
		return nil
	}
	runDaemonProcessFn = func(filteredArgs []string) error {
		daemonCalls++
		return nil
	}
	bootstrapHelperProcessFn = func(_ []string) error {
		return nil
	}

	exitCode := runEntry([]string{"SkillFlow"})

	assert.Equal(t, 0, exitCode)
	assert.Equal(t, 1, daemonCalls)
	assert.Equal(t, 0, runUICalls)
}

func TestBuildUIOptionsDoesNotHideWindowOnCloseWhenDaemonOwnsTrayLifecycle(t *testing.T) {
	opts := buildUIOptions(NewApp())
	require.NotNil(t, opts)
	assert.False(t, opts.HideWindowOnClose)
}
