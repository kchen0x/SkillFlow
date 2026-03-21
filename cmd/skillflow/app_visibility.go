package main

import "github.com/shinerio/skillflow/core/platform/eventbus"

func (a *App) publishWindowVisibilityChanged(visible bool) {
	if a == nil {
		return
	}

	a.windowVisibilityMu.Lock()
	if a.windowVisibilityInit && a.windowVisible == visible {
		a.windowVisibilityMu.Unlock()
		return
	}
	a.windowVisibilityInit = true
	a.windowVisible = visible
	a.windowVisibilityMu.Unlock()

	if a.hub == nil {
		return
	}

	a.hub.Publish(eventbus.Event{
		Type: eventbus.EventAppWindowVisibilityChanged,
		Payload: eventbus.AppWindowVisibilityPayload{
			Visible: visible,
		},
	})
}
