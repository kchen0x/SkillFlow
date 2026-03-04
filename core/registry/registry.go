package registry

import (
	"github.com/shinerio/skillflow/core/backup"
	"github.com/shinerio/skillflow/core/install"
	skillsync "github.com/shinerio/skillflow/core/sync"
)

var (
	installers     = map[string]install.Installer{}
	adapters       = map[string]skillsync.ToolAdapter{}
	cloudProviders = map[string]backup.CloudProvider{}
)

func RegisterInstaller(i install.Installer)       { installers[i.Type()] = i }
func RegisterAdapter(a skillsync.ToolAdapter)     { adapters[a.Name()] = a }
func RegisterCloudProvider(p backup.CloudProvider) { cloudProviders[p.Name()] = p }

func GetInstaller(t string) (install.Installer, bool) {
	i, ok := installers[t]
	return i, ok
}

func GetAdapter(name string) (skillsync.ToolAdapter, bool) {
	a, ok := adapters[name]
	return a, ok
}

func GetCloudProvider(name string) (backup.CloudProvider, bool) {
	p, ok := cloudProviders[name]
	return p, ok
}

func AllAdapters() []skillsync.ToolAdapter {
	result := make([]skillsync.ToolAdapter, 0, len(adapters))
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
