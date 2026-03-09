package backup

type ProviderFactory func() CloudProvider

var providerFactories []ProviderFactory

func RegisterProviderFactory(factory ProviderFactory) {
	providerFactories = append(providerFactories, factory)
}

func RegisteredProviders() []CloudProvider {
	providers := make([]CloudProvider, 0, len(providerFactories))
	for _, factory := range providerFactories {
		providers = append(providers, factory())
	}
	return providers
}
