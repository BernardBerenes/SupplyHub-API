package storage

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithyhttp "github.com/aws/smithy-go/transport/http"

	"github.com/BernardBerenes/SupplyHub-API/internal/config"
)

type MinIO struct {
	client *s3.Client
	bucket string
}

func NewMinIO(cfg *config.Config) (*MinIO, error) {
	client := s3.New(s3.Options{
		BaseEndpoint: aws.String(resolveEndpoint(cfg.MinIOEndpoint, cfg.MinIOUseSSL)),
		Region:       cfg.MinIORegion,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		UsePathStyle: true,
	})

	ctx := context.Background()

	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(cfg.MinIOBucket)})
	if err != nil {
		var respErr *smithyhttp.ResponseError
		if !errors.As(err, &respErr) || respErr.HTTPStatusCode() != 404 {
			return nil, err
		}

		if _, err := client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(cfg.MinIOBucket)}); err != nil {
			return nil, err
		}
	}

	return &MinIO{
		client: client,
		bucket: cfg.MinIOBucket,
	}, nil
}

func resolveEndpoint(endpoint string, useSSL bool) string {
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return endpoint
	}

	scheme := "http"
	if useSSL {
		scheme = "https"
	}

	return scheme + "://" + endpoint
}

func (m *MinIO) Upload(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error {
	_, err := m.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(m.bucket),
		Key:           aws.String(objectKey),
		Body:          reader,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})

	return err
}

func (m *MinIO) Delete(ctx context.Context, objectKey string) error {
	_, err := m.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(objectKey),
	})

	return err
}
