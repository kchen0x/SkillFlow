package domain

import "context"

const GitProviderName = "git"

type CredentialField struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Placeholder string `json:"placeholder"`
	Secret      bool   `json:"secret"`
}

type RemoteFile struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	IsDir  bool   `json:"isDir"`
	Action string `json:"action,omitempty"`
}

type Snapshot map[string]SnapshotEntry

type SnapshotEntry struct {
	Size int64  `json:"size"`
	Hash string `json:"hash"`
}

type BackupProfile struct {
	Provider         string
	BucketName       string
	RemotePath       string
	Credentials      map[string]string
	AppDataDir       string
}

type RunResult struct {
	Files    []RemoteFile
	Snapshot Snapshot
	RootDir  string
}

type CloudProvider interface {
	Name() string
	Init(credentials map[string]string) error
	Sync(ctx context.Context, localDir, bucket, remotePath string, onProgress func(file string)) error
	Restore(ctx context.Context, bucket, remotePath, localDir string) error
	List(ctx context.Context, bucket, remotePath string) ([]RemoteFile, error)
	RequiredCredentials() []CredentialField
}

type GitBackupProvider interface {
	CloudProvider
	PendingChanges(localDir string) ([]RemoteFile, error)
	ResolveConflictUseLocal(localDir string) error
	ResolveConflictUseRemote(localDir string) error
}

type GitConflictError struct {
	Output string
	Files  []string
}

func (e *GitConflictError) Error() string {
	return "git conflict: " + e.Output
}
