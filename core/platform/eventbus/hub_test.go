package eventbus_test

import (
	"testing"
	"time"

	"github.com/shinerio/skillflow/core/platform/eventbus"
	"github.com/stretchr/testify/assert"
)

func TestHubPublishSubscribe(t *testing.T) {
	hub := eventbus.NewHub()
	ch := hub.Subscribe()
	defer hub.Unsubscribe(ch)

	hub.Publish(eventbus.Event{Type: eventbus.EventBackupStarted})

	select {
	case evt := <-ch:
		assert.Equal(t, eventbus.EventBackupStarted, evt.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected event, got timeout")
	}
}

func TestHubMultipleSubscribers(t *testing.T) {
	hub := eventbus.NewHub()
	ch1 := hub.Subscribe()
	ch2 := hub.Subscribe()
	defer hub.Unsubscribe(ch1)
	defer hub.Unsubscribe(ch2)

	hub.Publish(eventbus.Event{Type: eventbus.EventSyncCompleted})

	for _, ch := range []<-chan eventbus.Event{ch1, ch2} {
		select {
		case evt := <-ch:
			assert.Equal(t, eventbus.EventSyncCompleted, evt.Type)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("subscriber did not receive event")
		}
	}
}
