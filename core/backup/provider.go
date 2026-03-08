package backup

import "context"

type CredentialField struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Placeholder string `json:"placeholder"`
	Secret      bool   `json:"secret"`
}

type RemoteFile struct {
	Path  string `json:"path"`
	Size  int64  `json:"size"`
	IsDir bool   `json:"isDir"`
	Action string `json:"action,omitempty"`
}

type CloudProvider interface {
	Name() string
	Init(credentials map[string]string) error
	// Sync mirrors localDir to cloud bucket at remotePath (incremental, no compression)
	Sync(ctx context.Context, localDir, bucket, remotePath string, onProgress func(file string)) error
	Restore(ctx context.Context, bucket, remotePath, localDir string) error
	List(ctx context.Context, bucket, remotePath string) ([]RemoteFile, error)
	RequiredCredentials() []CredentialField
}
