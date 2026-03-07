package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	obs "github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
)

type HuaweiProvider struct {
	client *obs.ObsClient
}

func NewHuaweiProvider() *HuaweiProvider { return &HuaweiProvider{} }

func (h *HuaweiProvider) Name() string { return "huawei" }

func (h *HuaweiProvider) RequiredCredentials() []CredentialField {
	return []CredentialField{
		{Key: "access_key_id", Label: "Access Key ID", Secret: false},
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
		if ShouldSkipBackupPath(rel) {
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
			if ShouldSkipBackupPath(rel) {
				continue
			}
			local := filepath.Join(localDir, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(local), 0755); err != nil {
				return err
			}
			getInput := &obs.GetObjectInput{}
			getInput.Bucket = bucketName
			getInput.Key = obj.Key
			resp, err := h.client.GetObject(getInput)
			if err != nil {
				return err
			}
			f, err := os.Create(local)
			if err != nil {
				resp.Body.Close()
				return err
			}
			_, err = f.ReadFrom(resp.Body)
			resp.Body.Close()
			f.Close()
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

func (h *HuaweiProvider) List(_ context.Context, bucket, remotePath string) ([]RemoteFile, error) {
	bucketName, err := h.bucketName(bucket)
	if err != nil {
		return nil, err
	}
	input := &obs.ListObjectsInput{Bucket: bucketName}
	input.Prefix = remotePath
	var files []RemoteFile
	for {
		result, err := h.client.ListObjects(input)
		if err != nil {
			return nil, err
		}
		for _, obj := range result.Contents {
			rel := strings.TrimPrefix(obj.Key, remotePath)
			if ShouldSkipBackupPath(rel) {
				continue
			}
			files = append(files, RemoteFile{
				Path: rel,
				Size: obj.Size,
			})
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
	bucket, err := normalizeHuaweiBucketName(raw)
	if err != nil {
		return "", err
	}
	return bucket, nil
}

func normalizeHuaweiEndpoint(raw string) (string, error) {
	endpoint := normalizeHostLikeValue(raw)
	if endpoint == "" {
		return "", fmt.Errorf("huawei obs endpoint is required")
	}
	parts := strings.Split(endpoint, ".")
	if len(parts) >= 5 && parts[1] == "obs" && isValidHuaweiBucketName(parts[0]) {
		endpoint = strings.Join(parts[1:], ".")
	}
	return endpoint, nil
}

func normalizeHuaweiBucketName(raw string) (string, error) {
	bucket := strings.TrimSpace(raw)
	if bucket == "" {
		return "", fmt.Errorf("huawei obs bucket name is required")
	}
	if isValidHuaweiBucketName(bucket) {
		return bucket, nil
	}
	host := normalizeHostLikeValue(bucket)
	if host != "" {
		parts := strings.Split(host, ".")
		if len(parts) >= 5 && parts[1] == "obs" && isValidHuaweiBucketName(parts[0]) {
			return parts[0], nil
		}
	}
	return "", fmt.Errorf("invalid huawei obs bucket name %q: enter bucket name only, not the full OBS URL or host", raw)
}

func isValidHuaweiBucketName(name string) bool {
	if len(name) < 3 || len(name) > 63 {
		return false
	}
	for i, r := range name {
		isLower := r >= 'a' && r <= 'z'
		isDigit := r >= '0' && r <= '9'
		isHyphen := r == '-'
		if !isLower && !isDigit && !isHyphen {
			return false
		}
		if (i == 0 || i == len(name)-1) && !isLower && !isDigit {
			return false
		}
	}
	return true
}
