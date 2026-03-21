package snapshot

import (
	"os"
	"path/filepath"
	"time"
)

func MigrateLegacyNestedGitDir(skillsDir, backupRoot string) (string, bool, error) {
	cleanSkillsDir := filepath.Clean(skillsDir)
	cleanBackupRoot := filepath.Clean(backupRoot)
	if cleanSkillsDir == "" || cleanBackupRoot == "" || cleanSkillsDir == cleanBackupRoot {
		return "", false, nil
	}

	legacyGitPath := filepath.Join(cleanSkillsDir, ".git")
	if _, err := os.Stat(legacyGitPath); err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}

	targetPath := filepath.Join(cleanSkillsDir, ".git.skillflow-legacy-backup")
	if _, err := os.Stat(targetPath); err == nil {
		targetPath = filepath.Join(cleanSkillsDir, ".git.skillflow-legacy-backup."+time.Now().Format("20060102150405"))
	} else if !os.IsNotExist(err) {
		return "", false, err
	}

	if err := os.Rename(legacyGitPath, targetPath); err != nil {
		return "", false, err
	}
	return targetPath, true, nil
}
