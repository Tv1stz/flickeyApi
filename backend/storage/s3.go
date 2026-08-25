// Package storage provides the S3-compatible object storage implementation.
// Compatible with Selectel Object Storage, AWS S3, and MinIO.
package storage

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"flickey/go-backend/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Storage implements StorageProvider using AWS SDK v2.
type S3Storage struct {
	client    *s3.Client
	presigner *s3.PresignClient
	bucket    string
}

// NewS3Storage creates an S3Storage configured for Selectel (or any S3-compatible endpoint).
func NewS3Storage(cfg *config.Settings) (*S3Storage, error) {
	endpoint := cfg.S3EndpointURL
	if endpoint != "" && !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}

	resolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...any) (aws.Endpoint, error) {
			if endpoint != "" {
				return aws.Endpoint{
					URL:               endpoint,
					HostnameImmutable: true,
					SigningRegion:     region,
				}, nil
			}
			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		},
	)

	awsCfg, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		awsconfig.WithRegion(cfg.S3Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		),
		awsconfig.WithEndpointResolverWithOptions(resolver), //nolint:staticcheck
	)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
		o.UsePathStyle = true
	})

	return &S3Storage{
		client:    client,
		presigner: s3.NewPresignClient(client),
		bucket:    cfg.S3Bucket,
	}, nil
}

// GeneratePresignedUploadURL generates a presigned PUT URL for direct client uploads.
func (s *S3Storage) GeneratePresignedUploadURL(ctx context.Context, fileKey, contentType string, expiresIn int) (string, error) {
	req, err := s.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(fileKey),
		ContentType: aws.String(contentType),
	}, func(o *s3.PresignOptions) {
		o.Expires = time.Duration(expiresIn) * time.Second
	})
	if err != nil {
		return "", fmt.Errorf("GeneratePresignedUploadURL: %w", err)
	}
	return req.URL, nil
}

// ObjectExists checks if an object exists via HEAD request.
func (s *S3Storage) ObjectExists(ctx context.Context, fileKey string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		if isS3NotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("ObjectExists: %w", err)
	}
	return true, nil
}

// GetObjectMetadata returns object metadata via HEAD.
func (s *S3Storage) GetObjectMetadata(ctx context.Context, fileKey string) (map[string]any, error) {
	out, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		if isS3NotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("GetObjectMetadata: %w", err)
	}

	meta := map[string]any{}
	if out.ContentLength != nil {
		meta["content_length"] = *out.ContentLength
	}
	if out.ContentType != nil {
		meta["content_type"] = *out.ContentType
	}
	if out.ETag != nil {
		meta["etag"] = *out.ETag
	}
	return meta, nil
}

// GetObjectBytes reads the first maxBytes bytes of an object (for magic byte inspection).
func (s *S3Storage) GetObjectBytes(ctx context.Context, fileKey string, maxBytes int) ([]byte, error) {
	rangeStr := fmt.Sprintf("bytes=0-%d", maxBytes-1)
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fileKey),
		Range:  aws.String(rangeStr),
	})
	if err != nil {
		if isS3NotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("GetObjectBytes: %w", err)
	}
	return ReadAll(out.Body, maxBytes)
}

// Delete removes an object from S3. Idempotent — ignores 404.
func (s *S3Storage) Delete(ctx context.Context, fileKey string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fileKey),
	})
	if err != nil && !isS3NotFound(err) {
		return fmt.Errorf("Delete: %w", err)
	}
	return nil
}

// PutObject stores raw bytes (dev upload handler compatibility).
func (s *S3Storage) PutObject(ctx context.Context, fileKey string, data []byte, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(fileKey),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("PutObject: %w", err)
	}
	return nil
}

func isS3NotFound(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "NoSuchKey") ||
		strings.Contains(msg, "NotFound") ||
		strings.Contains(msg, "404")
}
