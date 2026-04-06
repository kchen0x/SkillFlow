package main

import (
	"context"
	"encoding/json"

	"github.com/shinerio/skillflow/core/platform/eventbus"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	EventMemoryContentChanged = string(eventbus.EventMemoryContentChanged)
	EventMemoryPushCompleted  = string(eventbus.EventMemoryPushCompleted)
	EventMemoryStatusChanged  = string(eventbus.EventMemoryStatusChanged)
)

func forwardEvents(ctx context.Context, hub *eventbus.Hub) {
	ch := hub.Subscribe()
	for {
		select {
		case evt, ok := <-ch:
			if !ok {
				return
			}
			emitRuntimeEvent(ctx, evt)
		case <-ctx.Done():
			return
		}
	}
}

func emitRuntimeEvent(ctx context.Context, evt eventbus.Event) {
	data, _ := json.Marshal(evt.Payload)
	runtime.EventsEmit(ctx, string(evt.Type), string(data))
}
