package registry

import (
	"github.com/shinerio/skillflow/core/backup"
	skillsync "github.com/shinerio/skillflow/core/sync"
)

var (
	adapters       = map[string]skillsync.AgentAdapter{}
	cloudProviders = map[string]backup.CloudProvider{}
)

func RegisterAdapter(a skillsync.AgentAdapter)     { adapters[a.Name()] = a }
func RegisterCloudProvider(p backup.CloudProvider) { cloudProviders[p.Name()] = p }

func GetAdapter(name string) (skillsync.AgentAdapter, bool) {
	a, ok := adapters[name]
	return a, ok
}

func GetCloudProvider(name string) (backup.CloudProvider, bool) {
	p, ok := cloudProviders[name]
	return p, ok
}

func AllAdapters() []skillsync.AgentAdapter {
	result := make([]skillsync.AgentAdapter, 0, len(adapters))
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
