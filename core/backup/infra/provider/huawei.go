//go:build !provider_select || backup_huawei

package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	obs "github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	backupdomain "github.com/shinerio/skillflow/core/backup/domain"
	snapshotinfra "github.com/shinerio/skillflow/core/backup/infra/snapshot"
)

type HuaweiProvider struct {
	client *obs.ObsClient
}

func NewHuaweiProvider() *HuaweiProvider { return &HuaweiProvider{} }

func init() {
	RegisterProviderFactory(func() backupdomain.CloudProvider { return NewHuaweiProvider() })
}

func (h *HuaweiProvider) Name() string { return "huawei" }

func (h *HuaweiProvider) RequiredCredentials() []backupdomain.CredentialField {
	return []backupdomain.CredentialField{
		{Key: "access_key_id", Label: "Access Key ID"},
		{Key: "secret_access_key", Label: "Secret Access Key", Secret: true},
		{Key: "endpoint", Label: "Endpoint", Placeholder: "obs.cn-north-4.myhuaweicloud.com"},
	}
}

func (h *HuaweiProvider) Init(creds map[string]string) error {
	endpoint, err := normalizeHuaweiEndpoint(creds["endpoint"])
	if err != nil {
		return err
	}
	client, err := obs.New(creds["access_key_id"], creds["secret_access_key"], endpoint)
	if err != nil {
		return fmt.Errorf("init huawei obs client failed: %w", err)
	}
	h.client = client
	return nil
}

func (h *HuaweiProvider) Sync(_ context.Context, localDir, bucket, remotePath string, onProgress func(string)) error {
	bucketName, err := h.bucketName(bucket)
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
		input := &obs.PutFileInput{}
		input.Bucket = bucketName
		input.Key = key
		input.SourceFile = path
		_, err = h.client.PutFile(input)
		return err
	})
}

func (h *HuaweiProvider) Restore(_ context.Context, bucket, remotePath, localDir string) error {
	bucketName, err := h.bucketName(bucket)
	if err != nil {
		return err
	}
	input := &obs.ListObjectsInput{Bucket: bucketName}
	input.Prefix = remotePath
	for {
		result, err := h.client.ListObjects(input)
		if err != nil {
			return err
		}
		for _, obj := range result.Contents {
			rel := strings.TrimPrefix(obj.Key, remotePath)
			if snapshotinfra.ShouldSkipBackupPath(rel) {
				continue
			}
			local := filepath.Join(localDir, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(local), 0o755); err != nil {
				return err
			}
			getInput := &obs.GetObjectInput{}
			getInput.Bucket = bucketName
			getInput.Key = obj.Key
			resp, err := h.client.GetObject(getInput)
			if err != nil {
				return err
			}
			file, err := os.Create(local)
			if err != nil {
				resp.Body.Close()
				return err
			}
			_, err = file.ReadFrom(resp.Body)
			resp.Body.Close()
			file.Close()
			if err != nil {
				return err
			}
		}
		if !result.IsTruncated {
			break
		}
		input.Marker = result.NextMarker
	}
	return nil
}

func (h *HuaweiProvider) List(_ context.Context, bucket, remotePath string) ([]backupdomain.RemoteFile, error) {
	bucketName, err := h.bucketName(bucket)
	if err != nil {
		return nil, err
	}
	input := &obs.ListObjectsInput{Bucket: bucketName}
	input.Prefix = remotePath
	var files []backupdomain.RemoteFile
	for {
		result, err := h.client.ListObjects(input)
		if err != nil {
			return nil, err
		}
		for _, obj := range result.Contents {
			rel := strings.TrimPrefix(obj.Key, remotePath)
			if snapshotinfra.ShouldSkipBackupPath(rel) {
				continue
			}
			files = append(files, backupdomain.RemoteFile{Path: rel, Size: obj.Size})
		}
		if !result.IsTruncated {
			break
		}
		input.Marker = result.NextMarker
	}
	return files, nil
}

func (h *HuaweiProvider) bucketName(raw string) (string, error) {
	if h.client == nil {
		return "", fmt.Errorf("huawei obs client is not initialized")
	}
	return normalizeHuaweiBucketName(raw)
}

func normalizeHuaweiEndpoint(raw string) (string, error) {
	endpoint := normalizeHostLikeValue(raw)
	if endpoint == "" {
		return "", fmt.Errorf("huawei obs endpoint is required")
	}
	parts := strings.Split(endpoint, ".")
	if len(parts) >= 5 && parts[1] == "obs" && isValidBucketName(parts[0]) {
		endpoint = strings.Join(parts[1:], ".")
	}
	return endpoint, nil
}

func normalizeHuaweiBucketName(raw string) (string, error) {
	bucket := strings.TrimSpace(raw)
	if bucket == "" {
		return "", fmt.Errorf("huawei obs bucket name is required")
	}
	if isValidBucketName(bucket) {
		return bucket, nil
	}
	host := normalizeHostLikeValue(bucket)
	if host != "" {
		parts := strings.Split(host, ".")
		if len(parts) >= 5 && parts[1] == "obs" && isValidBucketName(parts[0]) {
			return parts[0], nil
		}
	}
	return "", invalidBucketNameError("huawei obs", raw)
}
