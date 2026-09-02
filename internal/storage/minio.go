package storage

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/BernardBerenes/SupplyHub-API/internal/config"
)

type MinIO struct {
	client *minio.Client
	bucket string
}

func NewMinIO(cfg *config.Config) (*MinIO, error) {
	client, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: cfg.MinIOUseSSL,
	})
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	exists, err := client.BucketExists(ctx, cfg.MinIOBucket)
	if err != nil {
		return nil, err
	}

	if !exists {
		if err := client.MakeBucket(ctx, cfg.MinIOBucket, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
	}

	return &MinIO{
		client: client,
		bucket: cfg.MinIOBucket,
	}, nil
}

func (m *MinIO) Upload(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error {
	_, err := m.client.PutObject(ctx, m.bucket, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})

	return err
}

func (m *MinIO) Delete(ctx context.Context, objectKey string) error {
	return m.client.RemoveObject(ctx, m.bucket, objectKey, minio.RemoveObjectOptions{})
}
