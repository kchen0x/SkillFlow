//go:build darwin

package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	darwinEnsureStatusItemFn        = ensureDarwinStatusItem
	darwinApplyRegularPolicyFn      = applyDarwinRegularPolicy
	darwinRuntimeShowFn             = runtime.Show
	darwinRuntimeWindowShowFn       = runtime.WindowShow
	darwinRuntimeWindowUnminimiseFn = runtime.WindowUnminimise
	darwinRuntimeHideFn             = runtime.Hide
)

func showMainWindowNative(ctx context.Context) error {
	darwinEnsureStatusItemFn()
	darwinApplyRegularPolicyFn()
	darwinRuntimeShowFn(ctx)
	darwinRuntimeWindowShowFn(ctx)
	darwinRuntimeWindowUnminimiseFn(ctx)
	return nil
}

func hideMainWindowNative(ctx context.Context) error {
	darwinEnsureStatusItemFn()
	darwinRuntimeHideFn(ctx)
	return nil
}
