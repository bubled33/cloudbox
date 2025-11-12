package storage

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	args := m.Called(ctx, params)
	return &s3.DeleteObjectOutput{}, args.Error(1)
}

func (m *MockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	args := m.Called(ctx, params)
	return &s3.PutObjectOutput{}, args.Error(1)
}

func (m *MockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	args := m.Called(ctx, params)

	body := io.NopCloser(bytes.NewReader([]byte("file content")))
	return &s3.GetObjectOutput{Body: body}, args.Error(1)
}

func (m *MockS3Client) HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	args := m.Called(ctx, params)
	return &s3.HeadObjectOutput{}, args.Error(1)
}

type MockS3Presigner struct {
	mock.Mock
}

func (m *MockS3Presigner) PresignPutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	args := m.Called(ctx, params)
	return &v4.PresignedHTTPRequest{URL: "http://upload-url"}, args.Error(1)
}

func (m *MockS3Presigner) PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	args := m.Called(ctx, params)
	return &v4.PresignedHTTPRequest{URL: "http://download-url"}, args.Error(1)
}

func TestGenerateUploadURL(t *testing.T) {
	ctx := context.Background()
	mockPresigner := new(MockS3Presigner)
	mockClient := new(MockS3Client)

	mockPresigner.
		On("PresignPutObject", ctx, mock.AnythingOfType("*s3.PutObjectInput")).
		Return(&v4.PresignedHTTPRequest{URL: "http://upload-url"}, nil)

	s := &S3Storage{
		client:        mockClient,
		presignClient: mockPresigner,
		bucket:        "test-bucket",
	}

	url, err := s.GenerateUploadURL(ctx, "test-key", 5*time.Minute)
	assert.NoError(t, err)
	assert.Equal(t, "http://upload-url", url)

	mockPresigner.AssertExpectations(t)
}

func TestGenerateDownloadURL(t *testing.T) {
	ctx := context.Background()
	mockPresigner := new(MockS3Presigner)
	mockClient := new(MockS3Client)

	mockPresigner.
		On("PresignGetObject", ctx, mock.AnythingOfType("*s3.GetObjectInput")).
		Return(&v4.PresignedHTTPRequest{URL: "http://download-url"}, nil)

	s := &S3Storage{
		client:        mockClient,
		presignClient: mockPresigner,
		bucket:        "test-bucket",
	}

	url, err := s.GenerateDownloadURL(ctx, "test-key", 5*time.Minute)
	assert.NoError(t, err)
	assert.Equal(t, "http://download-url", url)

	mockPresigner.AssertExpectations(t)
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	mockPresigner := new(MockS3Presigner)
	mockClient := new(MockS3Client)

	mockClient.
		On("DeleteObject", ctx, mock.AnythingOfType("*s3.DeleteObjectInput")).
		Return(&s3.DeleteObjectOutput{}, nil)

	s := &S3Storage{
		client:        mockClient,
		presignClient: mockPresigner,
		bucket:        "test-bucket",
	}

	err := s.Delete(ctx, "test-key")
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestUploadFile(t *testing.T) {
	ctx := context.Background()
	mockPresigner := new(MockS3Presigner)
	mockClient := new(MockS3Client)

	fileData := []byte("file content")

	mockClient.
		On("PutObject", ctx, mock.MatchedBy(func(input *s3.PutObjectInput) bool {
			return *input.Key == "test-key"
		})).
		Return(&s3.PutObjectOutput{}, nil)

	s := &S3Storage{
		client:        mockClient,
		presignClient: mockPresigner,
		bucket:        "test-bucket",
	}

	err := s.UploadFile(ctx, "test-key", fileData)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestDownloadFile(t *testing.T) {
	ctx := context.Background()
	mockPresigner := new(MockS3Presigner)
	mockClient := new(MockS3Client)

	mockClient.
		On("GetObject", ctx, mock.MatchedBy(func(input *s3.GetObjectInput) bool {
			return *input.Key == "test-key"
		})).
		Return(&s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader([]byte("file content")))}, nil)

	s := &S3Storage{
		client:        mockClient,
		presignClient: mockPresigner,
		bucket:        "test-bucket",
	}

	data, err := s.DownloadFile(ctx, "test-key")
	assert.NoError(t, err)
	assert.Equal(t, []byte("file content"), data)

	mockClient.AssertExpectations(t)
}
