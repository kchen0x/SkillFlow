package main

import (
	"context"

	agentapp "github.com/shinerio/skillflow/core/agentintegration/app"
	agentgateway "github.com/shinerio/skillflow/core/agentintegration/app/port/gateway"
	agentdomain "github.com/shinerio/skillflow/core/agentintegration/domain"
	agentgatewayinfra "github.com/shinerio/skillflow/core/agentintegration/infra/gateway"
	backupapp "github.com/shinerio/skillflow/core/backup/app"
	backupdomain "github.com/shinerio/skillflow/core/backup/domain"
	"github.com/shinerio/skillflow/core/config"
	platformgit "github.com/shinerio/skillflow/core/platform/git"
	skillsourceapp "github.com/shinerio/skillflow/core/skillsource/app"
	sourcediscovery "github.com/shinerio/skillflow/core/skillsource/infra/discovery"
)

var registeredAgentGateways = map[string]agentgateway.AgentGateway{}

func registerAdapters() {
	registeredAgentGateways = map[string]agentgateway.AgentGateway{}
	for _, name := range agentdomain.BuiltinAgentNames() {
		registerAgentGateway(agentgatewayinfra.NewFilesystemAdapter(name, agentdomain.DefaultProfile(name).PushDir))
	}
}

func registerAgentGateway(gateway agentgateway.AgentGateway) {
	registeredAgentGateways[gateway.Name()] = gateway
}

func agentGateway(name string) (agentgateway.AgentGateway, bool) {
	gateway, ok := registeredAgentGateways[name]
	return gateway, ok
}

func resolveAgentGateway(profile agentdomain.AgentProfile) agentgateway.AgentGateway {
	if adapter, ok := agentGateway(profile.Name); ok {
		return adapter
	}
	return agentgatewayinfra.NewFilesystemAdapter(profile.Name, profile.PushDir)
}

func newAgentIntegrationService() *agentapp.Service {
	return agentapp.NewService(resolveAgentGateway)
}

type appSkillsourceGitClient struct{}

func (appSkillsourceGitClient) CheckGitInstalled() error {
	return platformgit.CheckGitInstalled()
}

func (appSkillsourceGitClient) ParseRepoName(repoURL string) (string, error) {
	return platformgit.ParseRepoName(repoURL)
}

func (appSkillsourceGitClient) RepoSource(repoURL string) (string, error) {
	return platformgit.RepoSource(repoURL)
}

func (appSkillsourceGitClient) CacheDir(dataDir, repoURL string) (string, error) {
	return platformgit.CacheDir(dataDir, repoURL)
}

func (appSkillsourceGitClient) SameRepo(repoA, repoB string) bool {
	return platformgit.SameRepo(repoA, repoB)
}

func (appSkillsourceGitClient) CloneOrUpdate(ctx context.Context, repoURL, dir, proxyURL string) error {
	return cloneOrUpdateRepo(ctx, repoURL, dir, proxyURL)
}

func (appSkillsourceGitClient) CloneOrUpdateWithCreds(ctx context.Context, repoURL, dir, proxyURL, username, password string) error {
	return platformgit.CloneOrUpdateWithCreds(ctx, repoURL, dir, proxyURL, username, password)
}

func (appSkillsourceGitClient) IsAuthError(err error) bool {
	return platformgit.IsAuthError(err)
}

func (appSkillsourceGitClient) IsSSHAuthError(err error) bool {
	return platformgit.IsSSHAuthError(err)
}

func (a *App) newSkillsourceService() *skillsourceapp.Service {
	return skillsourceapp.NewService(a.starStorage, sourcediscovery.NewRepoScanner(), appSkillsourceGitClient{})
}

func resolveCloudProvider(name string) (backupdomain.CloudProvider, bool) {
	return cloudProvider(name)
}

func (a *App) newBackupService() *backupapp.Service {
	return backupapp.NewService(resolveCloudProvider)
}

func (a *App) backupProfile(cfg config.AppConfig) backupdomain.BackupProfile {
	return backupdomain.BackupProfile{
		Provider:         cfg.Cloud.Provider,
		BucketName:       cfg.Cloud.BucketName,
		RemotePath:       cfg.Cloud.RemotePath,
		Credentials:      cfg.Cloud.Credentials,
		SkillsStorageDir: cfg.SkillsStorageDir,
		AppDataDir:       config.AppDataDir(),
	}
}
