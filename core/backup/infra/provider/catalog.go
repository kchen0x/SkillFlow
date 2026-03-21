package provider

import backupdomain "github.com/shinerio/skillflow/core/backup/domain"

type Factory func() backupdomain.CloudProvider

var factories []Factory

func RegisterProviderFactory(factory Factory) {
	factories = append(factories, factory)
}

func RegisteredProviders() []backupdomain.CloudProvider {
	providers := make([]backupdomain.CloudProvider, 0, len(factories))
	for _, factory := range factories {
		providers = append(providers, factory())
	}
	return providers
}

func RegisteredProviderFactories() []Factory {
	return append([]Factory(nil), factories...)
}
