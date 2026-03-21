package main

import (
	backupprovider "github.com/shinerio/skillflow/core/backup/infra/provider"
	"github.com/shinerio/skillflow/core/registry"
)

func registerProviders() {
	for _, factory := range backupprovider.RegisteredProviderFactories() {
		registry.RegisterCloudProviderFactory(factory)
	}
}
