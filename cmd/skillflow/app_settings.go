package main

import "github.com/shinerio/skillflow/core/config"

func (a *App) GetConfig() (config.AppConfig, error) {
	var (
		cfg config.AppConfig
		err error
	)
	if shouldProxyAppMethodsToDaemon() {
		err = a.invokeDaemonService("GetConfig", nil, &cfg)
	} else {
		cfg, err = a.config.Load()
	}
	if err != nil {
		return cfg, err
	}
	cfg.DefaultCategory = defaultCategoryName
	cfg.LogLevel = config.NormalizeLogLevel(cfg.LogLevel)
	cfg.RepoScanMaxDepth = config.NormalizeRepoScanMaxDepth(cfg.RepoScanMaxDepth)
	return cfg, nil
}

func (a *App) GetAppDataDir() string {
	return a.dataDir()
}

func (a *App) OpenAppDataDir() error {
	appDataDir := a.dataDir()
	a.logInfof("open app data dir started: appDataDir=%s", appDataDir)
	if err := a.OpenPath(appDataDir); err != nil {
		a.logErrorf("open app data dir failed: appDataDir=%s err=%v", appDataDir, err)
		return err
	}
	a.logInfof("open app data dir completed: appDataDir=%s", appDataDir)
	return nil
}

func (a *App) SaveConfig(cfg config.AppConfig) error {
	return a.newSettingsSaveCoordinator().Save(cfg)
}

type settingsSaveCoordinator struct {
	app *App
}

func (a *App) newSettingsSaveCoordinator() *settingsSaveCoordinator {
	return &settingsSaveCoordinator{app: a}
}

func (c *settingsSaveCoordinator) Save(cfg config.AppConfig) error {
	c.app.logInfof("save config requested")
	prevCfg, err := c.app.config.Load()
	if err != nil {
		c.app.logErrorf("save config failed: load current config failed: %v", err)
		return err
	}
	cfg.DefaultCategory = defaultCategoryName
	cfg.LogLevel = config.NormalizeLogLevel(cfg.LogLevel)
	cfg.RepoScanMaxDepth = config.NormalizeRepoScanMaxDepth(cfg.RepoScanMaxDepth)
	if err := c.app.config.Save(cfg); err != nil {
		c.app.logErrorf("save config failed: %v", err)
		return err
	}
	c.app.rebuildPathBoundServices(cfg.RepoCacheDir)
	c.syncExistingSkillsForNewAutoPushAgents(prevCfg, cfg)
	if prevCfg.LaunchAtLogin != cfg.LaunchAtLogin {
		if err := c.app.syncLaunchAtLogin(cfg.LaunchAtLogin); err != nil {
			persistedCfg := cfg
			rollbackCfg := cfg
			rollbackCfg.LaunchAtLogin = prevCfg.LaunchAtLogin
			if rollbackErr := c.app.config.Save(rollbackCfg); rollbackErr != nil {
				c.app.logErrorf("save config rollback failed: restore launch-at-login setting=%t err=%v", prevCfg.LaunchAtLogin, rollbackErr)
			} else if rollbackErr := c.app.syncLaunchAtLogin(prevCfg.LaunchAtLogin); rollbackErr != nil {
				c.app.logErrorf("save config rollback failed: restore launch-at-login state=%t err=%v", prevCfg.LaunchAtLogin, rollbackErr)
				persistedCfg = rollbackCfg
			} else {
				persistedCfg = rollbackCfg
			}
			c.app.rebuildPathBoundServices(persistedCfg.RepoCacheDir)
			c.app.setLoggerLevel(persistedCfg.LogLevel)
			c.app.startAutoSyncTimer(persistedCfg.Cloud.SyncIntervalMinutes)
			c.app.logErrorf("save config failed: apply launch-at-login failed: %v", err)
			return err
		}
	}
	c.app.setLoggerLevel(cfg.LogLevel)
	c.app.logInfof("save config completed: logLevel=%s repoScanMaxDepth=%d launchAtLogin=%t", cfg.LogLevel, cfg.RepoScanMaxDepth, cfg.LaunchAtLogin)
	c.app.startAutoSyncTimer(cfg.Cloud.SyncIntervalMinutes)
	return nil
}

func (c *settingsSaveCoordinator) syncExistingSkillsForNewAutoPushAgents(prevCfg, nextCfg config.AppConfig) {
	prevTargets := make(map[string]struct{}, len(prevCfg.AutoPushAgents))
	for _, name := range prevCfg.AutoPushAgents {
		prevTargets[name] = struct{}{}
	}

	newTargets := make([]string, 0, len(nextCfg.AutoPushAgents))
	for _, name := range nextCfg.AutoPushAgents {
		if _, existed := prevTargets[name]; existed {
			continue
		}
		newTargets = append(newTargets, name)
	}
	if len(newTargets) == 0 {
		return
	}

	skills, err := c.app.storage.ListAll()
	if err != nil {
		c.app.logErrorf("sync existing skills to new auto push agents failed: load skills failed: %v", err)
		return
	}
	if len(skills) == 0 {
		c.app.logInfof("sync existing skills to new auto push agents skipped: reason=no-skills agentCount=%d", len(newTargets))
		return
	}

	skillIDs := make([]string, 0, len(skills))
	for _, sk := range skills {
		skillIDs = append(skillIDs, sk.ID)
	}

	c.app.logInfof("sync existing skills to new auto push agents started: skillCount=%d agentCount=%d", len(skillIDs), len(newTargets))
	conflicts, err := c.app.PushToAgents(skillIDs, newTargets)
	if err != nil {
		c.app.logErrorf("sync existing skills to new auto push agents failed: %v", err)
		return
	}
	c.app.logInfof("sync existing skills to new auto push agents completed: skillCount=%d agentCount=%d conflicts=%d", len(skillIDs), len(newTargets), len(conflicts))
}
