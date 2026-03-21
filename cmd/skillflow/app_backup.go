package main

import "github.com/shinerio/skillflow/core/config"

// OpenGitBackupDir opens the git backup root directory in the OS file manager.
func (a *App) OpenGitBackupDir() error {
	backupDir := a.newBackupService().BackupRootDir(a.backupProfile(config.AppConfig{}))
	a.logInfof("open git backup dir started: backupDir=%s", backupDir)
	if err := a.OpenPath(backupDir); err != nil {
		a.logErrorf("open git backup dir failed: backupDir=%s err=%v", backupDir, err)
		return err
	}
	a.logInfof("open git backup dir completed: backupDir=%s", backupDir)
	return nil
}
