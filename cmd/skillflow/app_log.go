package main

import (
	"fmt"
	"path/filepath"

	"github.com/shinerio/skillflow/core/platform/logging"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) logDir() string {
	return filepath.Join(appDataDirFunc(), "logs")
}

func (a *App) initLogger(logLevel string) {
	lg, err := logging.New(a.logDir(), logLevel)
	if err != nil {
		if a.ctx != nil {
			runtime.LogErrorf(a.ctx, "logger init failed: %v", err)
		}
		return
	}
	a.sysLog = lg
	a.logInfof("logger initialized, level=%s dir=%s", lg.LevelString(), lg.Dir())
}

func (a *App) setLoggerLevel(level string) string {
	normalized := logging.NormalizeLevelString(level)
	if a.sysLog != nil {
		a.sysLog.SetLevelString(normalized)
	}
	return normalized
}

func (a *App) logDebugf(format string, args ...any) {
	a.logf(logging.LevelDebug, format, args...)
}

func (a *App) logInfof(format string, args ...any) {
	a.logf(logging.LevelInfo, format, args...)
}

func (a *App) logErrorf(format string, args ...any) {
	a.logf(logging.LevelError, format, args...)
}

func (a *App) logf(level logging.Level, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if a.sysLog != nil {
		a.sysLog.Logf(level, "%s", msg)
		if !a.sysLog.Enabled(level) {
			return
		}
	}
	if a.ctx == nil {
		return
	}
	switch level {
	case logging.LevelDebug:
		runtime.LogDebug(a.ctx, msg)
	case logging.LevelError:
		runtime.LogError(a.ctx, msg)
	default:
		runtime.LogInfo(a.ctx, msg)
	}
}
