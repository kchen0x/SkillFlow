package app

import (
	"context"
	"errors"
	"testing"

	backupdomain "github.com/shinerio/skillflow/core/backup/domain"
)

type fakeProvider struct {
	name          string
	listFiles     []backupdomain.RemoteFile
	pendingFiles  []backupdomain.RemoteFile
	restoreErr    error
	syncErr       error
	listErr       error
	resolveLocal  error
	resolveRemote error
	inited        map[string]string
}

func (f *fakeProvider) Name() string { return f.name }
func (f *fakeProvider) Init(credentials map[string]string) error {
	f.inited = credentials
	return nil
}
func (f *fakeProvider) Sync(ctx context.Context, localDir, bucket, remotePath string, onProgress func(file string)) error {
	return f.syncErr
}
func (f *fakeProvider) Restore(ctx context.Context, bucket, remotePath, localDir string) error {
	return f.restoreErr
}
func (f *fakeProvider) List(ctx context.Context, bucket, remotePath string) ([]backupdomain.RemoteFile, error) {
	return f.listFiles, f.listErr
}
func (f *fakeProvider) RequiredCredentials() []backupdomain.CredentialField { return nil }
func (f *fakeProvider) PendingChanges(localDir string) ([]backupdomain.RemoteFile, error) {
	return f.pendingFiles, nil
}
func (f *fakeProvider) ResolveConflictUseLocal(localDir string) error {
	return f.resolveLocal
}
func (f *fakeProvider) ResolveConflictUseRemote(localDir string) error {
	return f.resolveRemote
}

func TestRunBackupUsesGitPendingChangesWithoutSnapshotPersistence(t *testing.T) {
	dataDir := t.TempDir()
	skillsDir := dataDir + "/skills"
	provider := &fakeProvider{
		name: backupdomain.GitProviderName,
		pendingFiles: []backupdomain.RemoteFile{
			{Path: "config.json", Action: "modified"},
		},
	}
	svc := NewService(func(name string) (backupdomain.CloudProvider, bool) {
		return provider, true
	})

	result, err := svc.RunBackup(context.Background(), backupdomain.BackupProfile{
		Provider:         backupdomain.GitProviderName,
		SkillsStorageDir: skillsDir,
		AppDataDir:       dataDir,
		Credentials:      map[string]string{"repo_url": "repo"},
	}, func(string) {})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Files) != 1 || result.Files[0].Path != "config.json" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestListRemoteBackupFilesDelegatesToProvider(t *testing.T) {
	provider := &fakeProvider{
		name: "mock",
		listFiles: []backupdomain.RemoteFile{
			{Path: "skills/demo/skill.md", Size: 10},
		},
	}
	svc := NewService(func(name string) (backupdomain.CloudProvider, bool) {
		return provider, true
	})

	files, err := svc.ListRemoteBackupFiles(context.Background(), backupdomain.BackupProfile{
		Provider:    "mock",
		BucketName:  "bucket",
		RemotePath:  "root",
		Credentials: map[string]string{"key": "value"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Path != "skills/demo/skill.md" {
		t.Fatalf("unexpected files: %+v", files)
	}
}

func TestRestoreBackupReturnsGitConflictError(t *testing.T) {
	dataDir := t.TempDir()
	skillsDir := dataDir + "/skills"
	conflict := &backupdomain.GitConflictError{Output: "conflict", Files: []string{"config.json"}}
	provider := &fakeProvider{name: backupdomain.GitProviderName, restoreErr: conflict}
	svc := NewService(func(name string) (backupdomain.CloudProvider, bool) {
		return provider, true
	})

	_, err := svc.RestoreBackup(context.Background(), backupdomain.BackupProfile{
		Provider:         backupdomain.GitProviderName,
		SkillsStorageDir: skillsDir,
		AppDataDir:       dataDir,
		Credentials:      map[string]string{"repo_url": "repo"},
	})
	if !errors.As(err, &conflict) {
		t.Fatalf("expected git conflict, got %v", err)
	}
}

func TestResolveGitConflictUsesRequestedStrategy(t *testing.T) {
	dataDir := t.TempDir()
	skillsDir := dataDir + "/skills"
	provider := &fakeProvider{name: backupdomain.GitProviderName}
	svc := NewService(func(name string) (backupdomain.CloudProvider, bool) {
		return provider, true
	})

	if _, err := svc.ResolveGitConflict(backupdomain.BackupProfile{
		Provider:         backupdomain.GitProviderName,
		SkillsStorageDir: skillsDir,
		AppDataDir:       dataDir,
		Credentials:      map[string]string{"repo_url": "repo"},
	}, true); err != nil {
		t.Fatal(err)
	}
}
