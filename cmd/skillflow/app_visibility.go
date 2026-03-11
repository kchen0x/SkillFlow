package main

import "github.com/shinerio/skillflow/core/notify"

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

	a.hub.Publish(notify.Event{
		Type: notify.EventAppWindowVisibilityChanged,
		Payload: notify.AppWindowVisibilityPayload{
			Visible: visible,
		},
	})
}
