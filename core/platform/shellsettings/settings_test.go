package shellsettings_test

import (
	"testing"

	"github.com/shinerio/skillflow/core/platform/logging"
	"github.com/shinerio/skillflow/core/platform/shellsettings"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeProxyConfig(t *testing.T) {
	normalized := shellsettings.NormalizeProxyConfig(shellsettings.ProxyConfig{
		Mode: " MANUAL ",
		URL:  " http://127.0.0.1:7890 ",
	})
	assert.Equal(t, shellsettings.ProxyModeManual, normalized.Mode)
	assert.Equal(t, "http://127.0.0.1:7890", normalized.URL)

	normalized = shellsettings.NormalizeProxyConfig(shellsettings.ProxyConfig{
		Mode: "invalid",
		URL:  " ",
	})
	assert.Equal(t, shellsettings.ProxyModeNone, normalized.Mode)
	assert.Equal(t, "", normalized.URL)
	assert.True(t, shellsettings.IsZeroProxyConfig(normalized))
}

func TestNormalizeSharedSettings(t *testing.T) {
	normalized := shellsettings.NormalizeSharedSettings(shellsettings.SharedSettings{
		LogLevel:             "bad-level",
		SkippedUpdateVersion: " v1.2.3 ",
	})
	assert.Equal(t, logging.DefaultLevelString, normalized.LogLevel)
	assert.Equal(t, "v1.2.3", normalized.SkippedUpdateVersion)
}

func TestNormalizeLocalSettingsDropsInvalidWindow(t *testing.T) {
	normalized := shellsettings.NormalizeLocalSettings(shellsettings.LocalSettings{
		Window: &shellsettings.WindowState{Width: 320, Height: 200},
	})
	assert.Nil(t, normalized.Window)
}
