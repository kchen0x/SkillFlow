package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/shinerio/skillflow/core/platform/eventbus"
	daemonipc "github.com/shinerio/skillflow/core/platform/ipc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDaemonServiceRoundTrip(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "daemon-service.json")

	type echoParams struct {
		Name string `json:"name"`
	}
	type echoResult struct {
		Message string `json:"message"`
	}

	svc, err := StartService(statePath, map[string]ServiceHandler{
		"echo": func(ctx context.Context, params json.RawMessage) (any, error) {
			var req echoParams
			require.NoError(t, json.Unmarshal(params, &req))
			return echoResult{Message: "hello " + req.Name}, nil
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = svc.Close()
	})

	var result echoResult
	require.NoError(t, InvokeService(statePath, "echo", echoParams{Name: "daemon"}, &result))
	assert.Equal(t, echoResult{Message: "hello daemon"}, result)
}

func TestDaemonServiceRejectsUnauthorizedToken(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "daemon-service.json")

	svc, err := StartService(statePath, map[string]ServiceHandler{
		"ping": func(ctx context.Context, params json.RawMessage) (any, error) {
			return map[string]bool{"ok": true}, nil
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = svc.Close()
	})

	endpoint, err := daemonipc.ReadEndpoint(statePath)
	require.NoError(t, err)
	endpoint.Token = "wrong-token"
	require.NoError(t, daemonipc.WriteEndpoint(statePath, endpoint))

	err = InvokeService(statePath, "ping", nil, nil)
	require.Error(t, err)
	assert.EqualError(t, err, "unauthorized")
}

func TestDaemonServiceReturnsHandlerError(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "daemon-service.json")

	svc, err := StartService(statePath, map[string]ServiceHandler{
		"fail": func(ctx context.Context, params json.RawMessage) (any, error) {
			return nil, errors.New("boom")
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = svc.Close()
	})

	err = InvokeService(statePath, "fail", nil, nil)
	require.Error(t, err)
	assert.EqualError(t, err, "boom")
}

func TestDaemonServiceStreamsEvents(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "daemon-service.json")

	svc, err := StartService(statePath, map[string]ServiceHandler{})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = svc.Close()
	})

	hub := eventbus.NewHub()
	svc.SetEventHub(hub)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	received := make(chan eventbus.Event, 1)
	go func() {
		_ = StreamEvents(statePath, ctx, func(evt eventbus.Event) {
			received <- evt
			cancel()
		})
	}()

	time.Sleep(50 * time.Millisecond)
	hub.Publish(eventbus.Event{Type: eventbus.EventSkillsUpdated, Payload: map[string]any{"count": 1}})

	select {
	case evt := <-received:
		assert.Equal(t, eventbus.EventSkillsUpdated, evt.Type)
		payload, ok := evt.Payload.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, float64(1), payload["count"])
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for streamed event")
	}
}
