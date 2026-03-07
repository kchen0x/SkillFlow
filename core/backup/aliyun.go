package backup

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type AliyunProvider struct {
	client *oss.Client
}

func NewAliyunProvider() *AliyunProvider { return &AliyunProvider{} }

func (a *AliyunProvider) Name() string { return "aliyun" }

func (a *AliyunProvider) RequiredCredentials() []CredentialField {
	return []CredentialField{
		{Key: "access_key_id", Label: "Access Key ID", Secret: false},
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
	b, err := a.bucketHandle(bucket)
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
		return b.PutObjectFromFile(key, path)
	})
}

func (a *AliyunProvider) Restore(_ context.Context, bucket, remotePath, localDir string) error {
	b, err := a.bucketHandle(bucket)
	if err != nil {
		return err
	}
	marker := ""
	for {
		result, err := b.ListObjects(oss.Prefix(remotePath), oss.Marker(marker))
		if err != nil {
			return err
		}
		for _, obj := range result.Objects {
			rel := strings.TrimPrefix(obj.Key, remotePath)
			if ShouldSkipBackupPath(rel) {
				continue
			}
			local := filepath.Join(localDir, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(local), 0755); err != nil {
				return err
			}
			if err := b.GetObjectToFile(obj.Key, local); err != nil {
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

func (a *AliyunProvider) List(_ context.Context, bucket, remotePath string) ([]RemoteFile, error) {
	b, err := a.bucketHandle(bucket)
	if err != nil {
		return nil, err
	}
	var files []RemoteFile
	marker := ""
	for {
		result, err := b.ListObjects(oss.Prefix(remotePath), oss.Marker(marker))
		if err != nil {
			return nil, err
		}
		for _, obj := range result.Objects {
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
	b, err := a.client.Bucket(name)
	if err != nil {
		return nil, fmt.Errorf("open aliyun bucket %q failed: %w", name, err)
	}
	return b, nil
}

func normalizeAliyunEndpoint(raw string) (string, error) {
	endpoint := normalizeHostLikeValue(raw)
	if endpoint == "" {
		return "", fmt.Errorf("aliyun endpoint is required")
	}
	parts := strings.Split(endpoint, ".")
	if len(parts) >= 4 && strings.HasPrefix(parts[1], "oss-") {
		endpoint = strings.Join(parts[1:], ".")
	}
	return endpoint, nil
}

func normalizeAliyunBucketName(raw string) (string, error) {
	bucket := strings.TrimSpace(raw)
	if bucket == "" {
		return "", fmt.Errorf("aliyun bucket name is required")
	}
	if isValidAliyunBucketName(bucket) {
		return bucket, nil
	}
	host := normalizeHostLikeValue(bucket)
	if host != "" {
		parts := strings.Split(host, ".")
		if len(parts) >= 4 && strings.HasPrefix(parts[1], "oss-") && isValidAliyunBucketName(parts[0]) {
			return parts[0], nil
		}
	}
	return "", fmt.Errorf("invalid aliyun bucket name %q: enter bucket name only, not the full OSS URL or host", raw)
}

func normalizeHostLikeValue(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	if strings.Contains(value, "://") {
		if parsed, err := url.Parse(value); err == nil && parsed.Host != "" {
			value = parsed.Host
		}
	}
	value = strings.TrimPrefix(value, "//")
	if slash := strings.Index(value, "/"); slash >= 0 {
		value = value[:slash]
	}
	if host, _, err := net.SplitHostPort(value); err == nil {
		value = host
	}
	return strings.TrimSuffix(strings.TrimSpace(value), ".")
}

func isValidAliyunBucketName(name string) bool {
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
