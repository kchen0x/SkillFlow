package registry

import (
	agentgateway "github.com/shinerio/skillflow/core/agentintegration/app/port/gateway"
	"github.com/shinerio/skillflow/core/backup"
)

var (
	adapters       = map[string]agentgateway.AgentGateway{}
	cloudProviders = map[string]backup.CloudProvider{}
)

func RegisterAdapter(a agentgateway.AgentGateway)  { adapters[a.Name()] = a }
func RegisterCloudProvider(p backup.CloudProvider) { cloudProviders[p.Name()] = p }

func GetAdapter(name string) (agentgateway.AgentGateway, bool) {
	a, ok := adapters[name]
	return a, ok
}

func GetCloudProvider(name string) (backup.CloudProvider, bool) {
	p, ok := cloudProviders[name]
	return p, ok
}

func AllAdapters() []agentgateway.AgentGateway {
	result := make([]agentgateway.AgentGateway, 0, len(adapters))
	for _, a := range adapters {
		result = append(result, a)
	}
	return result
}

func AllCloudProviders() []backup.CloudProvider {
	result := make([]backup.CloudProvider, 0, len(cloudProviders))
	for _, p := range cloudProviders {
		result = append(result, p)
	}
	return result
}
