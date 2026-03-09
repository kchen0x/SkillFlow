package main

import (
	"github.com/shinerio/skillflow/core/backup"
	"github.com/shinerio/skillflow/core/registry"
)

func registerProviders() {
	for _, provider := range backup.RegisteredProviders() {
		registry.RegisterCloudProvider(provider)
	}
}
