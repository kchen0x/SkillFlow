//go:build !provider_select || backup_azure

package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	backupdomain "github.com/shinerio/skillflow/core/backup/domain"
	snapshotinfra "github.com/shinerio/skillflow/core/backup/infra/snapshot"
)

type AzureProvider struct {
	client *azblob.Client
}

func NewAzureProvider() *AzureProvider { return &AzureProvider{} }

func init() {
	RegisterProviderFactory(func() backupdomain.CloudProvider { return NewAzureProvider() })
}

func (a *AzureProvider) Name() string { return "azure" }

func (a *AzureProvider) RequiredCredentials() []backupdomain.CredentialField {
	return []backupdomain.CredentialField{
		{Key: "account_name", Label: "Account Name", Placeholder: "myskillflowstorage"},
		{Key: "account_key", Label: "Account Key", Secret: true},
		{Key: "service_url", Label: "Service URL", Placeholder: "https://myskillflowstorage.blob.core.windows.net/"},
	}
}

func (a *AzureProvider) Init(creds map[string]string) error {
	accountName := normalizeAzureAccountName(creds["account_name"])
	if accountName == "" {
		accountName = accountNameFromAzureServiceURL(creds["service_url"])
	}
	if accountName == "" {
		return fmt.Errorf("azure blob account name is required")
	}
	serviceURL, err := normalizeAzureServiceURL(creds["service_url"], accountName)
	if err != nil {
		return err
	}
	cred, err := azblob.NewSharedKeyCredential(accountName, strings.TrimSpace(creds["account_key"]))
	if err != nil {
		return fmt.Errorf("init azure blob shared key credential failed: %w", err)
	}
	client, err := azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
	if err != nil {
		return fmt.Errorf("init azure blob client failed: %w", err)
	}
	a.client = client
	return nil
}

func (a *AzureProvider) Sync(ctx context.Context, localDir, bucket, remotePath string, onProgress func(string)) error {
	containerName, err := a.containerName(bucket)
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
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = a.client.UploadFile(ctx, containerName, key, file, nil)
		return err
	})
}

func (a *AzureProvider) Restore(ctx context.Context, bucket, remotePath, localDir string) error {
	containerName, err := a.containerName(bucket)
	if err != nil {
		return err
	}
	pager := a.client.NewListBlobsFlatPager(containerName, &azblob.ListBlobsFlatOptions{
		Prefix: &remotePath,
	})
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, blob := range page.Segment.BlobItems {
			if blob.Name == nil {
				continue
			}
			rel := strings.TrimPrefix(*blob.Name, remotePath)
			if rel == "" || snapshotinfra.ShouldSkipBackupPath(rel) {
				continue
			}
			local := filepath.Join(localDir, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(local), 0o755); err != nil {
				return err
			}
			file, err := os.Create(local)
			if err != nil {
				return err
			}
			if _, err := a.client.DownloadFile(ctx, containerName, *blob.Name, file, nil); err != nil {
				file.Close()
				return err
			}
			if err := file.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *AzureProvider) List(ctx context.Context, bucket, remotePath string) ([]backupdomain.RemoteFile, error) {
	containerName, err := a.containerName(bucket)
	if err != nil {
		return nil, err
	}
	pager := a.client.NewListBlobsFlatPager(containerName, &azblob.ListBlobsFlatOptions{
		Prefix: &remotePath,
	})
	var files []backupdomain.RemoteFile
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, blob := range page.Segment.BlobItems {
			if blob.Name == nil {
				continue
			}
			rel := strings.TrimPrefix(*blob.Name, remotePath)
			if rel == "" || snapshotinfra.ShouldSkipBackupPath(rel) {
				continue
			}
			var size int64
			if blob.Properties.ContentLength != nil {
				size = *blob.Properties.ContentLength
			}
			files = append(files, backupdomain.RemoteFile{Path: rel, Size: size})
		}
	}
	return files, nil
}

func (a *AzureProvider) containerName(raw string) (string, error) {
	if a.client == nil {
		return "", fmt.Errorf("azure blob client is not initialized")
	}
	container := strings.TrimSpace(raw)
	if container == "" {
		return "", fmt.Errorf("azure blob container name is required")
	}
	return container, nil
}

func normalizeAzureAccountName(raw string) string {
	return strings.TrimSpace(raw)
}

func normalizeAzureServiceURL(raw, accountName string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		if accountName == "" {
			return "", fmt.Errorf("azure blob service url is required")
		}
		return "https://" + accountName + ".blob.core.windows.net/", nil
	}
	if !strings.Contains(value, "://") {
		value = "https://" + value
	}
	if !strings.HasSuffix(value, "/") {
		value += "/"
	}
	return value, nil
}

func accountNameFromAzureServiceURL(raw string) string {
	host := normalizeHostLikeValue(raw)
	if host == "" {
		return ""
	}
	parts := strings.Split(host, ".")
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}
