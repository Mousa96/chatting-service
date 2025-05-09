package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

type S3Storage struct {
	client  S3Client
	bucket  string
	baseURL string
}

func NewS3Storage(client S3Client, bucket, baseURL string) *S3Storage {
	return &S3Storage{
		client:  client,
		bucket:  bucket,
		baseURL: baseURL,
	}
}

func (s *S3Storage) Upload(filename string, content io.Reader, contentType string) (string, error) {
	input := &s3.PutObjectInput{
		Bucket:      &s.bucket,
		Key:         &filename,
		Body:        content,
		ContentType: &contentType,
	}

	_, err := s.client.PutObject(context.Background(), input)
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	return fmt.Sprintf("%s/%s", s.baseURL, filename), nil
}

func (s *S3Storage) Delete(filename string) error {
	input := &s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    &filename,
	}

	_, err := s.client.DeleteObject(context.Background(), input)
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	return nil
} 