package main

import (
	backupdomain "github.com/shinerio/skillflow/core/backup/domain"
	backupprovider "github.com/shinerio/skillflow/core/backup/infra/provider"
)

var registeredCloudProviderFactories = map[string]func() backupdomain.CloudProvider{}

func registerProviders() {
	registeredCloudProviderFactories = map[string]func() backupdomain.CloudProvider{}
	for _, factory := range backupprovider.RegisteredProviderFactories() {
		registerCloudProviderFactory(factory)
	}
}

func registerCloudProviderFactory(factory func() backupdomain.CloudProvider) {
	provider := factory()
	registeredCloudProviderFactories[provider.Name()] = factory
}

func cloudProvider(name string) (backupdomain.CloudProvider, bool) {
	factory, ok := registeredCloudProviderFactories[name]
	if !ok {
		return nil, false
	}
	return factory(), true
}

func allCloudProviders() []backupdomain.CloudProvider {
	providers := make([]backupdomain.CloudProvider, 0, len(registeredCloudProviderFactories))
	for _, factory := range registeredCloudProviderFactories {
		providers = append(providers, factory())
	}
	return providers
}
