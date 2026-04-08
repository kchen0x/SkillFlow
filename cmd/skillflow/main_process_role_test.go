package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunEntryDefaultsToDaemon(t *testing.T) {
	prevDaemon := runDaemonProcessFn
	prevUI := runUIProcessFn
	t.Cleanup(func() {
		runDaemonProcessFn = prevDaemon
		runUIProcessFn = prevUI
	})

	calledDaemon := 0
	runDaemonProcessFn = func(filteredArgs []string) error {
		calledDaemon++
		return nil
	}
	runUIProcessFn = func() error {
		t.Fatal("runEntry should not call UI when daemon default role")
		return nil
	}

	assert.Equal(t, 0, runEntry([]string{"SkillFlow"}))
	assert.Equal(t, 1, calledDaemon)
}

func TestRunEntryInternalUIStartsUI(t *testing.T) {
	prevDaemon := runDaemonProcessFn
	prevUI := runUIProcessFn
	t.Cleanup(func() {
		runDaemonProcessFn = prevDaemon
		runUIProcessFn = prevUI
	})

	calledUI := 0
	runDaemonProcessFn = func(filteredArgs []string) error {
		t.Fatal("runEntry should not call daemon when --internal-ui is provided")
		return nil
	}
	runUIProcessFn = func() error {
		calledUI++
		return nil
	}

	assert.Equal(t, 0, runEntry([]string{"SkillFlow", internalUIFlag}))
	assert.Equal(t, 1, calledUI)
}

func TestRunEntryInternalDaemonStartsDaemon(t *testing.T) {
	prevDaemon := runDaemonProcessFn
	prevUI := runUIProcessFn
	t.Cleanup(func() {
		runDaemonProcessFn = prevDaemon
		runUIProcessFn = prevUI
	})

	var received []string
	runDaemonProcessFn = func(filteredArgs []string) error {
		received = append([]string(nil), filteredArgs...)
		return nil
	}
	runUIProcessFn = func() error {
		t.Fatal("runEntry should not call UI when --internal-daemon is provided")
		return nil
	}

	assert.Equal(t, 0, runEntry([]string{"SkillFlow", internalDaemonFlag, "--verbose"}))
	assert.Equal(t, []string{"SkillFlow", "--verbose"}, received)
}

func TestRunDaemonProcessDelegatesToBootstrapHelper(t *testing.T) {
	if !helperBootstrapEnabled() {
		t.Skip("helper bootstrap disabled in this build mode")
	}
	prevBootstrap := bootstrapHelperProcessFn
	prevUI := runUIProcessFn
	t.Cleanup(func() {
		bootstrapHelperProcessFn = prevBootstrap
		runUIProcessFn = prevUI
	})

	var received []string
	bootstrapHelperProcessFn = func(uiArgs []string) error {
		received = append([]string(nil), uiArgs...)
		return nil
	}
	runUIProcessFn = func() error {
		t.Fatal("runDaemonProcess should not start UI directly when helper bootstrap is enabled")
		return nil
	}

	assert.NoError(t, runDaemonProcess([]string{"SkillFlow", "--verbose"}))
	assert.Equal(t, []string{"--verbose"}, received)
}

func TestRunDaemonProcessStartsUIDirectlyWhenHelperBootstrapDisabled(t *testing.T) {
	prevBootstrap := bootstrapHelperProcessFn
	prevUI := runUIProcessFn
	t.Cleanup(func() {
		bootstrapHelperProcessFn = prevBootstrap
		runUIProcessFn = prevUI
	})

	helperCalled := false
	uiCalled := 0
	bootstrapHelperProcessFn = func(uiArgs []string) error {
		helperCalled = true
		return nil
	}
	runUIProcessFn = func() error {
		uiCalled++
		return nil
	}

	assert.NoError(t, runDaemonProcessWhenHelperDisabled([]string{"SkillFlow", "--verbose"}))
	assert.False(t, helperCalled)
	assert.Equal(t, 1, uiCalled)
}
