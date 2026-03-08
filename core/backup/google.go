package backup

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type GoogleProvider struct {
	client *storage.Client
}

func NewGoogleProvider() *GoogleProvider { return &GoogleProvider{} }

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
	if g.client != nil {
		_ = g.client.Close()
	}

	opt, err := googleCredentialOption(creds["service_account_json"])
	if err != nil {
		return err
	}

	client, err := storage.NewClient(context.Background(), opt)
	if err != nil {
		return fmt.Errorf("init google cloud storage client failed: %w", err)
	}

	g.client = client
	return nil
}

func (g *GoogleProvider) Sync(ctx context.Context, localDir, bucket, remotePath string, onProgress func(string)) error {
	bucketHandle, err := g.bucketHandle(bucket)
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

		writer := bucketHandle.Object(key).NewWriter(ctx)
		if _, err := io.Copy(writer, src); err != nil {
			_ = writer.Close()
			return err
		}
		return writer.Close()
	})
}

func (g *GoogleProvider) Restore(ctx context.Context, bucket, remotePath, localDir string) error {
	bucketHandle, err := g.bucketHandle(bucket)
	if err != nil {
		return err
	}

	it := bucketHandle.Objects(ctx, &storage.Query{Prefix: remotePath})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}

		rel := strings.TrimPrefix(attrs.Name, remotePath)
		if rel == "" || ShouldSkipBackupPath(rel) {
			continue
		}

		local := filepath.Join(localDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(local), 0755); err != nil {
			return err
		}

		reader, err := bucketHandle.Object(attrs.Name).NewReader(ctx)
		if err != nil {
			return err
		}
		if err := writeReaderToFile(local, reader); err != nil {
			return err
		}
	}

	return nil
}

func (g *GoogleProvider) List(ctx context.Context, bucket, remotePath string) ([]RemoteFile, error) {
	bucketHandle, err := g.bucketHandle(bucket)
	if err != nil {
		return nil, err
	}

	it := bucketHandle.Objects(ctx, &storage.Query{Prefix: remotePath})
	var files []RemoteFile
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		rel := strings.TrimPrefix(attrs.Name, remotePath)
		if rel == "" || ShouldSkipBackupPath(rel) {
			continue
		}
		files = append(files, RemoteFile{
			Path: rel,
			Size: attrs.Size,
		})
	}

	return files, nil
}

func (g *GoogleProvider) bucketHandle(bucket string) (*storage.BucketHandle, error) {
	if g.client == nil {
		return nil, fmt.Errorf("google cloud storage client is not initialized")
	}
	bucketName := strings.TrimSpace(bucket)
	if bucketName == "" {
		return nil, fmt.Errorf("google cloud storage bucket name is required")
	}
	return g.client.Bucket(bucketName), nil
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
