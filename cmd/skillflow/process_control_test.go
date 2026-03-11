package main

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetermineProcessRoleDefaultsToHelper(t *testing.T) {
	role, args := determineProcessRole([]string{"SkillFlow"})
	assert.Equal(t, processRoleHelper, role)
	assert.Equal(t, []string{"SkillFlow"}, args)
}

func TestDetermineProcessRoleReturnsUIAndStripsInternalFlag(t *testing.T) {
	role, args := determineProcessRole([]string{"SkillFlow", internalUIFlag, "--debug"})
	assert.Equal(t, processRoleUI, role)
	assert.Equal(t, []string{"SkillFlow", "--debug"}, args)
}

func TestLoopbackControlServerRoundTrip(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "helper-control.json")
	received := make(chan string, 1)

	server, err := startLoopbackControlServer(statePath, func(command string) error {
		received <- command
		return nil
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = server.Close()
	})

	require.NoError(t, sendLoopbackControlCommand(statePath, controlCommandShowUI))

	select {
	case command := <-received:
		assert.Equal(t, controlCommandShowUI, command)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for control command")
	}
}
