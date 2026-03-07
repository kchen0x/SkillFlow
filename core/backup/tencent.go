package backup

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	cos "github.com/tencentyun/cos-go-sdk-v5"
)

type TencentProvider struct {
	endpoint  string
	secretID  string
	secretKey string
}

func NewTencentProvider() *TencentProvider { return &TencentProvider{} }

func (t *TencentProvider) Name() string { return "tencent" }

func (t *TencentProvider) RequiredCredentials() []CredentialField {
	return []CredentialField{
		{Key: "secret_id", Label: "Secret ID", Secret: false},
		{Key: "secret_key", Label: "Secret Key", Secret: true},
		{Key: "endpoint", Label: "Endpoint", Placeholder: "cos.ap-guangzhou.myqcloud.com"},
	}
}

func (t *TencentProvider) Init(creds map[string]string) error {
	endpoint, err := normalizeTencentEndpoint(creds["endpoint"])
	if err != nil {
		return err
	}
	t.endpoint = endpoint
	t.secretID = creds["secret_id"]
	t.secretKey = creds["secret_key"]
	return nil
}

func (t *TencentProvider) Sync(ctx context.Context, localDir, bucket, remotePath string, onProgress func(string)) error {
	client, err := t.clientForBucket(bucket)
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
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = client.Object.Put(ctx, key, f, nil)
		return err
	})
}

func (t *TencentProvider) Restore(ctx context.Context, bucket, remotePath, localDir string) error {
	client, err := t.clientForBucket(bucket)
	if err != nil {
		return err
	}
	var marker string
	for {
		result, _, err := client.Bucket.Get(ctx, &cos.BucketGetOptions{
			Prefix: remotePath,
			Marker: marker,
		})
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
			_, err := client.Object.GetToFile(ctx, obj.Key, local, nil)
			if err != nil {
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

func (t *TencentProvider) List(ctx context.Context, bucket, remotePath string) ([]RemoteFile, error) {
	client, err := t.clientForBucket(bucket)
	if err != nil {
		return nil, err
	}
	var files []RemoteFile
	var marker string
	for {
		result, _, err := client.Bucket.Get(ctx, &cos.BucketGetOptions{
			Prefix: remotePath,
			Marker: marker,
		})
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
		marker = result.NextMarker
	}
	return files, nil
}

func (t *TencentProvider) clientForBucket(bucket string) (*cos.Client, error) {
	if strings.TrimSpace(t.endpoint) == "" {
		return nil, fmt.Errorf("tencent cos client is not initialized")
	}
	bucketURL, err := buildTencentBucketURL(bucket, t.endpoint)
	if err != nil {
		return nil, err
	}
	return cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  t.secretID,
			SecretKey: t.secretKey,
		},
	}), nil
}

func normalizeTencentEndpoint(raw string) (string, error) {
	endpoint := normalizeHostLikeValue(raw)
	if endpoint == "" {
		return "", fmt.Errorf("tencent cos endpoint is required")
	}
	parts := strings.Split(endpoint, ".")
	if len(parts) >= 5 && parts[1] == "cos" && isValidTencentBucketName(parts[0]) {
		endpoint = strings.Join(parts[1:], ".")
	}
	return endpoint, nil
}

func normalizeTencentBucketName(raw string) (string, error) {
	bucket := strings.TrimSpace(raw)
	if bucket == "" {
		return "", fmt.Errorf("tencent cos bucket name is required")
	}
	if isValidTencentBucketName(bucket) {
		return bucket, nil
	}
	host := normalizeHostLikeValue(bucket)
	if host != "" {
		parts := strings.Split(host, ".")
		if len(parts) >= 5 && parts[1] == "cos" && isValidTencentBucketName(parts[0]) {
			return parts[0], nil
		}
	}
	return "", fmt.Errorf("invalid tencent cos bucket name %q: enter bucket name only, not the full COS URL or host", raw)
}

func buildTencentBucketURL(bucket, endpoint string) (*url.URL, error) {
	bucketName, err := normalizeTencentBucketName(bucket)
	if err != nil {
		return nil, err
	}
	normalizedEndpoint, err := normalizeTencentEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse("https://" + bucketName + "." + normalizedEndpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid tencent cos bucket URL: %w", err)
	}
	return u, nil
}

func isValidTencentBucketName(name string) bool {
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
