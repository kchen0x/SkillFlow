package ipc

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

func TestLoopbackServerRoundTrip(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "daemon-control.json")
	received := make(chan string, 1)

	server, err := StartLoopbackServer(statePath, func(command string) error {
		received <- command
		return nil
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = server.Close()
	})

	require.NoError(t, SendLoopbackCommand(statePath, "show-ui"))

	select {
	case command := <-received:
		assert.Equal(t, "show-ui", command)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for control command")
	}
}

func TestSendLoopbackCommandRejectsUnauthorizedToken(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "daemon-control.json")

	server, err := StartLoopbackServer(statePath, func(command string) error {
		return nil
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = server.Close()
	})

	endpoint, err := ReadEndpoint(statePath)
	require.NoError(t, err)
	endpoint.Token = "wrong-token"
	require.NoError(t, writeEndpoint(statePath, endpoint))

	err = SendLoopbackCommand(statePath, "show-ui")
	require.Error(t, err)
	assert.EqualError(t, err, "unauthorized")
}

func TestPruneStaleStateRemovesDeadEndpoint(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "daemon-control.json")

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	address := listener.Addr().String()
	require.NoError(t, listener.Close())

	payload, err := json.Marshal(Endpoint{
		Address: address,
		Token:   "dead-token",
		PID:     999999,
	})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(statePath, payload, 0o644))

	require.NoError(t, PruneStaleState(statePath))
	_, err = os.Stat(statePath)
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestPruneStaleStateKeepsLiveEndpoint(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "daemon-control.json")

	server, err := StartLoopbackServer(statePath, func(command string) error {
		return nil
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = server.Close()
	})

	require.NoError(t, PruneStaleState(statePath))
	_, err = os.Stat(statePath)
	require.NoError(t, err)
}
