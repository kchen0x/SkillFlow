//go:build !provider_select || backup_google

package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/api/option"
	gcsapi "google.golang.org/api/storage/v1"
)

type GoogleProvider struct {
	service *gcsapi.Service
}

func NewGoogleProvider() *GoogleProvider { return &GoogleProvider{} }

func init() {
	RegisterProviderFactory(func() CloudProvider { return NewGoogleProvider() })
}

func (g *GoogleProvider) Name() string { return "google" }

func (g *GoogleProvider) RequiredCredentials() []CredentialField {
	return []CredentialField{
		{
			Key:         "service_account_json",
			Label:       "Service Account JSON",
			Placeholder: "Paste service account JSON or a local key file path",
			Secret:      true,
		},
	}
}

func (g *GoogleProvider) Init(creds map[string]string) error {
	opt, err := googleCredentialOption(creds["service_account_json"])
	if err != nil {
		return err
	}

	service, err := gcsapi.NewService(
		context.Background(),
		opt,
		option.WithScopes(gcsapi.DevstorageFullControlScope),
	)
	if err != nil {
		return fmt.Errorf("init google cloud storage service failed: %w", err)
	}

	g.service = service
	return nil
}

func (g *GoogleProvider) Sync(ctx context.Context, localDir, bucket, remotePath string, onProgress func(string)) error {
	bucketName, err := g.bucketName(bucket)
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

		src, err := os.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		_, err = g.service.Objects.Insert(bucketName, &gcsapi.Object{Name: key}).Context(ctx).Media(src).Do()
		return err
	})
}

func (g *GoogleProvider) Restore(ctx context.Context, bucket, remotePath, localDir string) error {
	bucketName, err := g.bucketName(bucket)
	if err != nil {
		return err
	}

	listCall := g.service.Objects.List(bucketName).Prefix(remotePath)
	return listCall.Pages(ctx, func(objects *gcsapi.Objects) error {
		for _, object := range objects.Items {
			if object == nil {
				continue
			}

			rel := strings.TrimPrefix(object.Name, remotePath)
			if rel == "" || ShouldSkipBackupPath(rel) {
				continue
			}

			local := filepath.Join(localDir, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(local), 0755); err != nil {
				return err
			}

			resp, err := g.service.Objects.Get(bucketName, object.Name).Context(ctx).Download()
			if err != nil {
				return err
			}
			if err := writeReaderToFile(local, resp.Body); err != nil {
				return err
			}
		}
		return nil
	})
}

func (g *GoogleProvider) List(ctx context.Context, bucket, remotePath string) ([]RemoteFile, error) {
	bucketName, err := g.bucketName(bucket)
	if err != nil {
		return nil, err
	}

	var files []RemoteFile
	listCall := g.service.Objects.List(bucketName).Prefix(remotePath)
	if err := listCall.Pages(ctx, func(objects *gcsapi.Objects) error {
		for _, object := range objects.Items {
			if object == nil {
				continue
			}

			rel := strings.TrimPrefix(object.Name, remotePath)
			if rel == "" || ShouldSkipBackupPath(rel) {
				continue
			}
			files = append(files, RemoteFile{
				Path: rel,
				Size: int64(object.Size),
			})
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return files, nil
}

func (g *GoogleProvider) bucketName(bucket string) (string, error) {
	if g.service == nil {
		return "", fmt.Errorf("google cloud storage service is not initialized")
	}
	bucketName := strings.TrimSpace(bucket)
	if bucketName == "" {
		return "", fmt.Errorf("google cloud storage bucket name is required")
	}
	return bucketName, nil
}

func googleCredentialOption(raw string) (option.ClientOption, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, fmt.Errorf("google cloud service account json is required")
	}
	if strings.HasPrefix(value, "{") {
		return option.WithCredentialsJSON([]byte(value)), nil
	}

	data, err := os.ReadFile(value)
	if err != nil {
		return nil, fmt.Errorf("read google cloud service account file %q failed: %w", value, err)
	}
	return option.WithCredentialsJSON(data), nil
}
