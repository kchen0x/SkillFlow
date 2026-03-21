package main

import (
	"testing"

	backupdomain "github.com/shinerio/skillflow/core/backup/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterProvidersIncludesGitProvider(t *testing.T) {
	registerProviders()

	provider, ok := cloudProvider(backupdomain.GitProviderName)
	require.True(t, ok)
	require.NotNil(t, provider)
	assert.Equal(t, backupdomain.GitProviderName, provider.Name())

	found := false
	for _, registered := range allCloudProviders() {
		if registered.Name() == backupdomain.GitProviderName {
			found = true
			break
		}
	}
	assert.True(t, found)
}
