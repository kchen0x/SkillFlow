package main

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetermineProcessRoleDefaultsToDaemon(t *testing.T) {
	role, args := determineProcessRole([]string{"SkillFlow"})
	assert.Equal(t, processRoleDaemon, role)
	assert.Equal(t, []string{"SkillFlow"}, args)
}

func TestDetermineProcessRoleFlags(t *testing.T) {
	roleDaemon, argsDaemon := determineProcessRole([]string{"SkillFlow", internalDaemonFlag, "--debug"})
	assert.Equal(t, processRoleDaemon, roleDaemon)
	assert.Equal(t, []string{"SkillFlow", "--debug"}, argsDaemon)

	roleUI, argsUI := determineProcessRole([]string{"SkillFlow", internalUIFlag, "--debug"})
	assert.Equal(t, processRoleUI, roleUI)
	assert.Equal(t, []string{"SkillFlow", "--debug"}, argsUI)

	roleLastWins, _ := determineProcessRole([]string{"SkillFlow", internalDaemonFlag, internalUIFlag})
	assert.Equal(t, processRoleUI, roleLastWins)
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

func TestPruneStaleLoopbackControlStateRemovesDeadEndpoint(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "helper-control.json")

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	address := listener.Addr().String()
	require.NoError(t, listener.Close())

	payload, err := json.Marshal(controlEndpoint{
		Address: address,
		Token:   "dead-token",
		PID:     999999,
	})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(statePath, payload, 0644))

	require.NoError(t, pruneStaleLoopbackControlState(statePath))
	_, err = os.Stat(statePath)
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestPruneStaleLoopbackControlStateKeepsLiveEndpoint(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "helper-control.json")

	server, err := startLoopbackControlServer(statePath, func(command string) error {
		return nil
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = server.Close()
	})

	require.NoError(t, pruneStaleLoopbackControlState(statePath))
	_, err = os.Stat(statePath)
	require.NoError(t, err)
}
