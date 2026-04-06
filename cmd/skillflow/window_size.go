package main

import (
	"context"

	"github.com/shinerio/skillflow/core/config"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	preferredWindowWidth  = 1460
	preferredWindowHeight = 920
	fallbackWindowWidth   = 1360
	fallbackWindowHeight  = 860
	minWindowWidth        = 960
	minWindowHeight       = 680
	windowWidthMargin     = 80
	windowHeightMargin    = 96
)

var windowGetSizeFn = runtime.WindowGetSize

func (a *App) fitInitialWindowToScreen(ctx context.Context) {
	a.logInfof("initial window sizing started")

	screens, err := runtime.ScreenGetAll(ctx)
	if err != nil {
		a.logErrorf("initial window sizing failed: screen query failed: %v", err)
		runtime.WindowSetMinSize(ctx, minWindowWidth, minWindowHeight)
		runtime.WindowSetSize(ctx, fallbackWindowWidth, fallbackWindowHeight)
		runtime.WindowCenter(ctx)
		a.logInfof("initial window sizing completed: fallback window=%dx%d min=%dx%d", fallbackWindowWidth, fallbackWindowHeight, minWindowWidth, minWindowHeight)
		return
	}

	screen := pickWindowSizingScreen(screens)
	screenWidth := screen.Size.Width
	screenHeight := screen.Size.Height
	if screenWidth <= 0 || screenHeight <= 0 {
		a.logErrorf("initial window sizing failed: invalid screen size width=%d height=%d", screenWidth, screenHeight)
		runtime.WindowSetMinSize(ctx, minWindowWidth, minWindowHeight)
		runtime.WindowSetSize(ctx, fallbackWindowWidth, fallbackWindowHeight)
		runtime.WindowCenter(ctx)
		a.logInfof("initial window sizing completed: fallback window=%dx%d min=%dx%d", fallbackWindowWidth, fallbackWindowHeight, minWindowWidth, minWindowHeight)
		return
	}

	targetWidth := clampInitialWindowDimension(preferredWindowWidth, screenWidth-windowWidthMargin)
	targetHeight := clampInitialWindowDimension(preferredWindowHeight, screenHeight-windowHeightMargin)
	if persisted, ok := a.config.LoadWindowState(); ok {
		targetWidth = clampInitialWindowDimension(persisted.Width, screenWidth-windowWidthMargin)
		targetHeight = clampInitialWindowDimension(persisted.Height, screenHeight-windowHeightMargin)
	}
	minWidth := minInt(minWindowWidth, targetWidth)
	minHeight := minInt(minWindowHeight, targetHeight)

	runtime.WindowSetMinSize(ctx, minWidth, minHeight)
	runtime.WindowSetSize(ctx, targetWidth, targetHeight)
	runtime.WindowCenter(ctx)
	a.initialWindowState = config.NormalizeWindowState(config.WindowState{Width: targetWidth, Height: targetHeight})

	a.logInfof(
		"initial window sizing completed: screen=%dx%d window=%dx%d min=%dx%d",
		screenWidth,
		screenHeight,
		targetWidth,
		targetHeight,
		minWidth,
		minHeight,
	)
}

func pickWindowSizingScreen(screens []runtime.Screen) runtime.Screen {
	for _, screen := range screens {
		if screen.IsCurrent {
			return screen
		}
	}
	for _, screen := range screens {
		if screen.IsPrimary {
			return screen
		}
	}
	if len(screens) > 0 {
		return screens[0]
	}
	return runtime.Screen{}
}

func clampInitialWindowDimension(preferred, available int) int {
	if available <= 0 {
		return preferred
	}
	if preferred > available {
		return available
	}
	return preferred
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (a *App) persistCurrentWindowSize(ctx context.Context) {
	a.logInfof("window size persistence started")
	width, height := windowGetSizeFn(ctx)
	state := config.NormalizeWindowState(config.WindowState{Width: width, Height: height})
	if state.Width == 0 || state.Height == 0 {
		a.logErrorf("window size persistence failed: invalid size width=%d height=%d", width, height)
		return
	}
	if state == a.initialWindowState {
		a.logDebugf("window size persistence skipped: unchanged width=%d height=%d", state.Width, state.Height)
		return
	}
	if err := a.config.SaveWindowState(state); err != nil {
		a.logErrorf("window size persistence failed: width=%d height=%d err=%v", state.Width, state.Height, err)
		return
	}
	a.logInfof("window size persistence completed: width=%d height=%d", state.Width, state.Height)
}
