package app

import (
	"context"
	"fmt"
	"path/filepath"

	backupdomain "github.com/shinerio/skillflow/core/backup/domain"
	snapshotinfra "github.com/shinerio/skillflow/core/backup/infra/snapshot"
	"github.com/shinerio/skillflow/core/platform/appdata"
)

type ProviderResolver func(name string) (backupdomain.CloudProvider, bool)

type Service struct {
	resolveProvider ProviderResolver
}

func NewService(resolveProvider ProviderResolver) *Service {
	return &Service{resolveProvider: resolveProvider}
}

func (s *Service) RunBackup(ctx context.Context, profile backupdomain.BackupProfile, onProgress func(string)) (backupdomain.RunResult, error) {
	provider, err := s.provider(profile)
	if err != nil {
		return backupdomain.RunResult{}, err
	}
	rootDir, err := s.prepareRoot(profile)
	if err != nil {
		return backupdomain.RunResult{}, err
	}
	if gitProvider, ok := provider.(backupdomain.GitBackupProvider); ok && profile.Provider == backupdomain.GitProviderName {
		files, err := gitProvider.PendingChanges(rootDir)
		if err != nil {
			return backupdomain.RunResult{}, err
		}
		if err := provider.Sync(ctx, rootDir, profile.BucketName, profile.RemotePath, onProgress); err != nil {
			return backupdomain.RunResult{}, err
		}
		return backupdomain.RunResult{Files: files, RootDir: rootDir}, nil
	}

	previousSnapshot, err := snapshotinfra.LoadSnapshot(s.snapshotPath(profile))
	if err != nil {
		return backupdomain.RunResult{}, err
	}
	currentSnapshot, err := snapshotinfra.BuildSnapshot(rootDir)
	if err != nil {
		return backupdomain.RunResult{}, err
	}
	if err := provider.Sync(ctx, rootDir, profile.BucketName, profile.RemotePath, onProgress); err != nil {
		return backupdomain.RunResult{}, err
	}
	if err := snapshotinfra.SaveSnapshot(s.snapshotPath(profile), currentSnapshot); err != nil {
		return backupdomain.RunResult{}, err
	}
	return backupdomain.RunResult{
		Files:    snapshotinfra.DiffSnapshots(previousSnapshot, currentSnapshot),
		Snapshot: currentSnapshot,
		RootDir:  rootDir,
	}, nil
}

func (s *Service) ListRemoteBackupFiles(ctx context.Context, profile backupdomain.BackupProfile) ([]backupdomain.RemoteFile, error) {
	provider, err := s.provider(profile)
	if err != nil {
		return nil, err
	}
	return provider.List(ctx, profile.BucketName, profile.RemotePath)
}

func (s *Service) RestoreBackup(ctx context.Context, profile backupdomain.BackupProfile) (backupdomain.RunResult, error) {
	provider, err := s.provider(profile)
	if err != nil {
		return backupdomain.RunResult{}, err
	}
	rootDir, err := s.prepareRoot(profile)
	if err != nil {
		return backupdomain.RunResult{}, err
	}
	beforeSnapshot, err := snapshotinfra.BuildSnapshot(rootDir)
	if err != nil {
		return backupdomain.RunResult{}, err
	}
	if err := provider.Restore(ctx, profile.BucketName, profile.RemotePath, rootDir); err != nil {
		return backupdomain.RunResult{}, err
	}
	afterSnapshot, err := snapshotinfra.BuildSnapshot(rootDir)
	if err != nil {
		return backupdomain.RunResult{}, err
	}
	if err := snapshotinfra.SaveSnapshot(s.snapshotPath(profile), afterSnapshot); err != nil {
		return backupdomain.RunResult{}, err
	}
	return backupdomain.RunResult{
		Files:    snapshotinfra.DiffSnapshots(beforeSnapshot, afterSnapshot),
		Snapshot: afterSnapshot,
		RootDir:  rootDir,
	}, nil
}

func (s *Service) ResolveGitConflict(profile backupdomain.BackupProfile, useLocal bool) (backupdomain.RunResult, error) {
	provider, err := s.provider(profile)
	if err != nil {
		return backupdomain.RunResult{}, err
	}
	gitProvider, ok := provider.(backupdomain.GitBackupProvider)
	if !ok {
		return backupdomain.RunResult{}, fmt.Errorf("provider %s does not support git conflict resolution", profile.Provider)
	}
	rootDir, err := s.prepareRoot(profile)
	if err != nil {
		return backupdomain.RunResult{}, err
	}
	beforeSnapshot, err := snapshotinfra.BuildSnapshot(rootDir)
	if err != nil {
		return backupdomain.RunResult{}, err
	}
	if useLocal {
		err = gitProvider.ResolveConflictUseLocal(rootDir)
	} else {
		err = gitProvider.ResolveConflictUseRemote(rootDir)
	}
	if err != nil {
		return backupdomain.RunResult{}, err
	}
	afterSnapshot, err := snapshotinfra.BuildSnapshot(rootDir)
	if err != nil {
		return backupdomain.RunResult{}, err
	}
	if err := snapshotinfra.SaveSnapshot(s.snapshotPath(profile), afterSnapshot); err != nil {
		return backupdomain.RunResult{}, err
	}
	return backupdomain.RunResult{
		Files:    snapshotinfra.DiffSnapshots(beforeSnapshot, afterSnapshot),
		Snapshot: afterSnapshot,
		RootDir:  rootDir,
	}, nil
}

func (s *Service) BackupRootDir(profile backupdomain.BackupProfile) string {
	return filepath.Clean(profile.AppDataDir)
}

func (s *Service) prepareRoot(profile backupdomain.BackupProfile) (string, error) {
	root := s.BackupRootDir(profile)
	if profile.Provider != backupdomain.GitProviderName {
		return root, nil
	}
	_, _, err := snapshotinfra.MigrateLegacyNestedGitDir(appdata.SkillsDir(root), root)
	if err != nil {
		return "", err
	}
	return root, nil
}

func (s *Service) snapshotPath(profile backupdomain.BackupProfile) string {
	return filepath.Join(profile.AppDataDir, "cache", "backup_snapshot.json")
}

func (s *Service) provider(profile backupdomain.BackupProfile) (backupdomain.CloudProvider, error) {
	provider, ok := s.resolveProvider(profile.Provider)
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", profile.Provider)
	}
	if err := provider.Init(profile.Credentials); err != nil {
		return nil, err
	}
	return provider, nil
}
