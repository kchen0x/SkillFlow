//go:build !provider_select || backup_aliyun

package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	backupdomain "github.com/shinerio/skillflow/core/backup/domain"
	snapshotinfra "github.com/shinerio/skillflow/core/backup/infra/snapshot"
)

type AliyunProvider struct {
	client *oss.Client
}

func NewAliyunProvider() *AliyunProvider { return &AliyunProvider{} }

func init() {
	RegisterProviderFactory(func() backupdomain.CloudProvider { return NewAliyunProvider() })
}

func (a *AliyunProvider) Name() string { return "aliyun" }

func (a *AliyunProvider) RequiredCredentials() []backupdomain.CredentialField {
	return []backupdomain.CredentialField{
		{Key: "access_key_id", Label: "Access Key ID"},
		{Key: "access_key_secret", Label: "Access Key Secret", Secret: true},
		{Key: "endpoint", Label: "Endpoint", Placeholder: "oss-cn-hangzhou.aliyuncs.com"},
	}
}

func (a *AliyunProvider) Init(creds map[string]string) error {
	endpoint, err := normalizeAliyunEndpoint(creds["endpoint"])
	if err != nil {
		return err
	}
	client, err := oss.New(endpoint, creds["access_key_id"], creds["access_key_secret"])
	if err != nil {
		return fmt.Errorf("init aliyun oss client failed: %w", err)
	}
	a.client = client
	return nil
}

func (a *AliyunProvider) Sync(_ context.Context, localDir, bucket, remotePath string, onProgress func(string)) error {
	bucketHandle, err := a.bucketHandle(bucket)
	if err != nil {
		return err
	}
	return filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(localDir, path)
		if snapshotinfra.ShouldSkipBackupPath(rel) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			return nil
		}
		key := remotePath + strings.ReplaceAll(rel, string(filepath.Separator), "/")
		if onProgress != nil {
			onProgress(rel)
		}
		return bucketHandle.PutObjectFromFile(key, path)
	})
}

func (a *AliyunProvider) Restore(_ context.Context, bucket, remotePath, localDir string) error {
	bucketHandle, err := a.bucketHandle(bucket)
	if err != nil {
		return err
	}
	marker := ""
	for {
		result, err := bucketHandle.ListObjects(oss.Prefix(remotePath), oss.Marker(marker))
		if err != nil {
			return err
		}
		for _, obj := range result.Objects {
			rel := strings.TrimPrefix(obj.Key, remotePath)
			if snapshotinfra.ShouldSkipBackupPath(rel) {
				continue
			}
			local := filepath.Join(localDir, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(local), 0o755); err != nil {
				return err
			}
			if err := bucketHandle.GetObjectToFile(obj.Key, local); err != nil {
				return err
			}
		}
		if !result.IsTruncated {
			break
		}
		marker = result.NextMarker
	}
	return nil
}

func (a *AliyunProvider) List(_ context.Context, bucket, remotePath string) ([]backupdomain.RemoteFile, error) {
	bucketHandle, err := a.bucketHandle(bucket)
	if err != nil {
		return nil, err
	}
	var files []backupdomain.RemoteFile
	marker := ""
	for {
		result, err := bucketHandle.ListObjects(oss.Prefix(remotePath), oss.Marker(marker))
		if err != nil {
			return nil, err
		}
		for _, obj := range result.Objects {
			rel := strings.TrimPrefix(obj.Key, remotePath)
			if snapshotinfra.ShouldSkipBackupPath(rel) {
				continue
			}
			files = append(files, backupdomain.RemoteFile{Path: rel, Size: obj.Size})
		}
		if !result.IsTruncated {
			break
		}
		marker = result.NextMarker
	}
	return files, nil
}

func (a *AliyunProvider) bucketHandle(bucket string) (*oss.Bucket, error) {
	if a.client == nil {
		return nil, fmt.Errorf("aliyun oss client is not initialized")
	}
	name, err := normalizeAliyunBucketName(bucket)
	if err != nil {
		return nil, err
	}
	handle, err := a.client.Bucket(name)
	if err != nil {
		return nil, fmt.Errorf("open aliyun bucket %q failed: %w", name, err)
	}
	return handle, nil
}

func normalizeAliyunEndpoint(raw string) (string, error) {
	endpoint := normalizeHostLikeValue(raw)
	if endpoint == "" {
		return "", fmt.Errorf("aliyun endpoint is required")
	}
	parts := strings.Split(endpoint, ".")
	if len(parts) >= 4 && strings.HasPrefix(parts[1], "oss-") && isValidBucketName(parts[0]) {
		endpoint = strings.Join(parts[1:], ".")
	}
	return endpoint, nil
}

func normalizeAliyunBucketName(raw string) (string, error) {
	bucket := strings.TrimSpace(raw)
	if bucket == "" {
		return "", fmt.Errorf("aliyun bucket name is required")
	}
	if isValidBucketName(bucket) {
		return bucket, nil
	}
	host := normalizeHostLikeValue(bucket)
	if host != "" {
		parts := strings.Split(host, ".")
		if len(parts) >= 4 && strings.HasPrefix(parts[1], "oss-") && isValidBucketName(parts[0]) {
			return parts[0], nil
		}
	}
	return "", invalidBucketNameError("aliyun oss", raw)
}
