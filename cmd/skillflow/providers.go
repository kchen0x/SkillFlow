package main

import (
	"github.com/shinerio/skillflow/core/backup"
	"github.com/shinerio/skillflow/core/registry"
)

func registerProviders() {
	registry.RegisterCloudProvider(backup.NewAliyunProvider())
	registry.RegisterCloudProvider(backup.NewAWSProvider())
	registry.RegisterCloudProvider(backup.NewGoogleProvider())
	registry.RegisterCloudProvider(backup.NewAzureProvider())
	registry.RegisterCloudProvider(backup.NewTencentProvider())
	registry.RegisterCloudProvider(backup.NewHuaweiProvider())
	registry.RegisterCloudProvider(backup.NewGitProvider())
}
