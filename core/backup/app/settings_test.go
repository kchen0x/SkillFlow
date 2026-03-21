package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultSettings(t *testing.T) {
	settings := DefaultSettings()
	assert.Equal(t, DefaultCloudRemotePath, settings.Cloud.RemotePath)
}

func TestNormalizeCloudRemotePath(t *testing.T) {
	assert.Equal(t, "team-a/nightly/skillflow/", NormalizeCloudRemotePath("team-a/nightly"))
	assert.Equal(t, DefaultCloudRemotePath, NormalizeCloudRemotePath(""))
	assert.Equal(t, DefaultCloudRemotePath, NormalizeCloudRemotePath("skillflow/"))
}
