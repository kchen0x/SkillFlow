package config_test

import (
	"testing"

	"github.com/shinerio/skillflow/core/config"
	"github.com/shinerio/skillflow/core/platform/appdata"
	"github.com/shinerio/skillflow/core/platform/shellsettings"
	"github.com/stretchr/testify/assert"
)

func TestAppDataDirDelegatesToPlatformAppData(t *testing.T) {
	assert.Equal(t, appdata.Dir(), config.AppDataDir())
}

func TestNormalizeProxyConfigDelegatesToShellSettings(t *testing.T) {
	input := config.ProxyConfig{Mode: config.ProxyModeManual, URL: " http://127.0.0.1:7890 "}
	assert.Equal(t, shellsettings.NormalizeProxyConfig(input), config.NormalizeProxyConfig(input))
}
