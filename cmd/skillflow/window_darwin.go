//go:build darwin

package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func showMainWindowNative(ctx context.Context) error {
	ensureDarwinStatusItem()
	applyDarwinRegularPolicy()
	runtime.Show(ctx)
	runtime.WindowShow(ctx)
	runtime.WindowUnminimise(ctx)
	return nil
}

func hideMainWindowNative(ctx context.Context) error {
	ensureDarwinStatusItem()
	runtime.Hide(ctx)
	return nil
}
