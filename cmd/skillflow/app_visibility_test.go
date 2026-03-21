package main

import (
	"testing"
	"time"

	"github.com/shinerio/skillflow/core/platform/eventbus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWindowVisibilityPublishesOnlyOnStateChange(t *testing.T) {
	app := NewApp()
	app.hub = eventbus.NewHub()

	events := app.hub.Subscribe()
	t.Cleanup(func() {
		app.hub.Unsubscribe(events)
	})

	app.publishWindowVisibilityChanged(false)
	first := readVisibilityEvent(t, events)
	assert.Equal(t, eventbus.EventAppWindowVisibilityChanged, first.Type)
	assert.Equal(t, eventbus.AppWindowVisibilityPayload{Visible: false}, first.Payload)

	app.publishWindowVisibilityChanged(false)
	assertNoVisibilityEvent(t, events)

	app.publishWindowVisibilityChanged(true)
	second := readVisibilityEvent(t, events)
	assert.Equal(t, eventbus.EventAppWindowVisibilityChanged, second.Type)
	assert.Equal(t, eventbus.AppWindowVisibilityPayload{Visible: true}, second.Payload)
}

func readVisibilityEvent(t *testing.T, events <-chan eventbus.Event) eventbus.Event {
	t.Helper()
	select {
	case evt := <-events:
		return evt
	case <-time.After(200 * time.Millisecond):
		require.FailNow(t, "expected visibility event")
		return eventbus.Event{}
	}
}

func assertNoVisibilityEvent(t *testing.T, events <-chan eventbus.Event) {
	t.Helper()
	select {
	case evt := <-events:
		require.FailNowf(t, "unexpected visibility event", "event=%+v", evt)
	case <-time.After(50 * time.Millisecond):
	}
}
