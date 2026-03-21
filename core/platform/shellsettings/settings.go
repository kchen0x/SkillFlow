package shellsettings

import (
	"strings"

	"github.com/shinerio/skillflow/core/platform/logging"
	"github.com/shinerio/skillflow/core/platform/settingsstore"
)

type WindowState = settingsstore.WindowState

type SharedSettings struct {
	LogLevel             string `json:"logLevel"`
	SkippedUpdateVersion string `json:"skippedUpdateVersion,omitempty"`
}

type LocalSettings struct {
	LaunchAtLogin bool        `json:"launchAtLogin"`
	Proxy         ProxyConfig `json:"proxy"`
	Window        *WindowState
}

func DefaultSharedSettings() SharedSettings {
	return SharedSettings{
		LogLevel: logging.DefaultLevelString,
	}
}

func DefaultLocalSettings() LocalSettings {
	return LocalSettings{
		Proxy: ProxyConfig{Mode: ProxyModeNone},
	}
}

func NormalizeSharedSettings(settings SharedSettings) SharedSettings {
	settings.LogLevel = logging.NormalizeLevelString(settings.LogLevel)
	settings.SkippedUpdateVersion = strings.TrimSpace(settings.SkippedUpdateVersion)
	return settings
}

func NormalizeLocalSettings(settings LocalSettings) LocalSettings {
	settings.Proxy = NormalizeProxyConfig(settings.Proxy)
	if settings.Window != nil {
		normalized := settingsstore.NormalizeWindowState(*settings.Window)
		if normalized.Width == 0 || normalized.Height == 0 {
			settings.Window = nil
		} else {
			settings.Window = &normalized
		}
	}
	return settings
}

func NormalizeWindowState(state WindowState) WindowState {
	return settingsstore.NormalizeWindowState(state)
}
