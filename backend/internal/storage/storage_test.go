package storage

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock S3 client
type mockS3Client struct {
    mock.Mock
}

func (m *mockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
    args := m.Called(ctx, params, optFns)
    return args.Get(0).(*s3.PutObjectOutput), args.Error(1)
}

func (m *mockS3Client) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
    args := m.Called(ctx, params, optFns)
    return args.Get(0).(*s3.DeleteObjectOutput), args.Error(1)
}

func TestLocalStorage(t *testing.T) {
    // Setup
    tmpDir := t.TempDir()
    storage := NewLocalStorage(tmpDir, "http://localhost:8080/uploads")

    tests := []struct {
        name        string
        filename    string
        content     []byte
        contentType string
        wantErr     bool
    }{
        {
            name:        "Valid file upload",
            filename:    "test.jpg",
            content:     []byte("fake image content"),
            contentType: "image/jpeg",
            wantErr:     false,
        },
        {
            name:        "Empty file",
            filename:    "empty.txt",
            content:     []byte{},
            contentType: "text/plain",
            wantErr:     false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test Upload
            reader := bytes.NewReader(tt.content)
            url, err := storage.Upload(tt.filename, reader, tt.contentType)

            if tt.wantErr {
                assert.Error(t, err)
                return
            }

            assert.NoError(t, err)
            assert.Contains(t, url, tt.filename)

            // Verify file exists and content
            content, err := os.ReadFile(tmpDir + "/" + tt.filename)
            assert.NoError(t, err)
            assert.Equal(t, tt.content, content)

            // Test Delete
            err = storage.Delete(tt.filename)
            assert.NoError(t, err)

            // Verify file is deleted
            _, err = os.Stat(tmpDir + "/" + tt.filename)
            assert.True(t, os.IsNotExist(err))
        })
    }
}

func TestS3Storage(t *testing.T) {
    mockClient := new(mockS3Client)
    storage := NewS3Storage(mockClient, "test-bucket", "https://test-bucket.s3.amazonaws.com")

    tests := []struct {
        name        string
        filename    string
        content     []byte
        contentType string
        wantErr     bool
    }{
        {
            name:        "Valid file upload",
            filename:    "test.jpg",
            content:     []byte("fake image content"),
            contentType: "image/jpeg",
            wantErr:     false,
        },
        {
            name:        "Empty file",
            filename:    "empty.txt",
            content:     []byte{},
            contentType: "text/plain",
            wantErr:     false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup expectations
            mockClient.On("PutObject", mock.Anything, mock.MatchedBy(func(input *s3.PutObjectInput) bool {
                return *input.Bucket == "test-bucket" && *input.Key == tt.filename
            }), mock.Anything).Return(&s3.PutObjectOutput{}, nil)

            mockClient.On("DeleteObject", mock.Anything, mock.MatchedBy(func(input *s3.DeleteObjectInput) bool {
                return *input.Bucket == "test-bucket" && *input.Key == tt.filename
            }), mock.Anything).Return(&s3.DeleteObjectOutput{}, nil)

            // Test Upload
            reader := bytes.NewReader(tt.content)
            url, err := storage.Upload(tt.filename, reader, tt.contentType)

            if tt.wantErr {
                assert.Error(t, err)
                return
            }

            assert.NoError(t, err)
            assert.Contains(t, url, tt.filename)

            // Test Delete
            err = storage.Delete(tt.filename)
            assert.NoError(t, err)

            mockClient.AssertExpectations(t)
        })
    }
} 