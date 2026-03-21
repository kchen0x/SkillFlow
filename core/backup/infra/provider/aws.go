//go:build !provider_select || backup_aws

package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	backupdomain "github.com/shinerio/skillflow/core/backup/domain"
	snapshotinfra "github.com/shinerio/skillflow/core/backup/infra/snapshot"
)

type AWSProvider struct {
	client *s3.Client
	region string
}

func NewAWSProvider() *AWSProvider { return &AWSProvider{} }

func init() {
	RegisterProviderFactory(func() backupdomain.CloudProvider { return NewAWSProvider() })
}

func (a *AWSProvider) Name() string { return "aws" }

func (a *AWSProvider) RequiredCredentials() []backupdomain.CredentialField {
	return []backupdomain.CredentialField{
		{Key: "access_key_id", Label: "Access Key ID"},
		{Key: "secret_access_key", Label: "Secret Access Key", Secret: true},
		{Key: "region", Label: "Region", Placeholder: "us-east-1"},
	}
}

func (a *AWSProvider) Init(creds map[string]string) error {
	region := normalizeAWSRegion(creds["region"])
	if region == "" {
		return fmt.Errorf("aws s3 region is required")
	}
	cfg := aws.Config{
		Region: region,
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
			strings.TrimSpace(creds["access_key_id"]),
			strings.TrimSpace(creds["secret_access_key"]),
			"",
		)),
	}
	a.client = s3.NewFromConfig(cfg)
	a.region = region
	return nil
}

func (a *AWSProvider) Sync(ctx context.Context, localDir, bucket, remotePath string, onProgress func(string)) error {
	bucketName, err := a.bucketName(bucket)
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
		_, err = a.client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
			Body:   file,
		})
		return err
	})
}

func (a *AWSProvider) Restore(ctx context.Context, bucket, remotePath, localDir string) error {
	bucketName, err := a.bucketName(bucket)
	if err != nil {
		return err
	}
	pager := s3.NewListObjectsV2Paginator(a.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(remotePath),
	})
	for pager.HasMorePages() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return err
		}
		for _, obj := range page.Contents {
			if obj.Key == nil {
				continue
			}
			rel := strings.TrimPrefix(aws.ToString(obj.Key), remotePath)
			if rel == "" || snapshotinfra.ShouldSkipBackupPath(rel) {
				continue
			}
			local := filepath.Join(localDir, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(local), 0o755); err != nil {
				return err
			}
			resp, err := a.client.GetObject(ctx, &s3.GetObjectInput{
				Bucket: aws.String(bucketName),
				Key:    obj.Key,
			})
			if err != nil {
				return err
			}
			if err := writeReaderToFile(local, resp.Body); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *AWSProvider) List(ctx context.Context, bucket, remotePath string) ([]backupdomain.RemoteFile, error) {
	bucketName, err := a.bucketName(bucket)
	if err != nil {
		return nil, err
	}
	pager := s3.NewListObjectsV2Paginator(a.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(remotePath),
	})
	var files []backupdomain.RemoteFile
	for pager.HasMorePages() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, obj := range page.Contents {
			if obj.Key == nil {
				continue
			}
			rel := strings.TrimPrefix(aws.ToString(obj.Key), remotePath)
			if rel == "" || snapshotinfra.ShouldSkipBackupPath(rel) {
				continue
			}
			var size int64
			if obj.Size != nil {
				size = *obj.Size
			}
			files = append(files, backupdomain.RemoteFile{Path: rel, Size: size})
		}
	}
	return files, nil
}

func (a *AWSProvider) bucketName(raw string) (string, error) {
	if a.client == nil {
		return "", fmt.Errorf("aws s3 client is not initialized")
	}
	bucket := strings.TrimSpace(raw)
	if bucket == "" {
		return "", fmt.Errorf("aws s3 bucket name is required")
	}
	return bucket, nil
}

func normalizeAWSRegion(raw string) string {
	return strings.TrimSpace(raw)
}
