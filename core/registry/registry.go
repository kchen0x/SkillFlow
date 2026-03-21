package registry

import (
	agentgateway "github.com/shinerio/skillflow/core/agentintegration/app/port/gateway"
	backupdomain "github.com/shinerio/skillflow/core/backup/domain"
)

var (
	adapters               = map[string]agentgateway.AgentGateway{}
	cloudProviderFactories = map[string]func() backupdomain.CloudProvider{}
)

func RegisterAdapter(a agentgateway.AgentGateway) { adapters[a.Name()] = a }

func RegisterCloudProviderFactory(factory func() backupdomain.CloudProvider) {
	provider := factory()
	cloudProviderFactories[provider.Name()] = factory
}

func GetAdapter(name string) (agentgateway.AgentGateway, bool) {
	a, ok := adapters[name]
	return a, ok
}

func GetCloudProvider(name string) (backupdomain.CloudProvider, bool) {
	factory, ok := cloudProviderFactories[name]
	if !ok {
		return nil, false
	}
	return factory(), true
}

func AllAdapters() []agentgateway.AgentGateway {
	result := make([]agentgateway.AgentGateway, 0, len(adapters))
	for _, a := range adapters {
		result = append(result, a)
	}
	return result
}

func AllCloudProviders() []backupdomain.CloudProvider {
	result := make([]backupdomain.CloudProvider, 0, len(cloudProviderFactories))
	for _, factory := range cloudProviderFactories {
		result = append(result, factory())
	}
	return result
}
