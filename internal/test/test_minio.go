package test

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	testMinio "github.com/testcontainers/testcontainers-go/modules/minio"
)

type TestS3Storage struct {
	container *testMinio.MinioContainer
	client    *minio.Client
	bucket    string
}

func SetupTestS3(ctx context.Context) (*TestS3Storage, error) {
	minioContainer, err := testMinio.Run(ctx,
		"minio/minio:RELEASE.2024-01-16T16-07-38Z",
		testMinio.WithUsername("minioadmin"),
		testMinio.WithPassword("minioadmin"),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to start minio: %w", err)
	}

	endpoint, err := minioContainer.ConnectionString(ctx)
	if err != nil {
		return nil, err
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4("minioadmin", "minioadmin", ""),
		Secure: false,
	})

	if err != nil {
		return nil, err
	}

	bucketName := "test-bucket"
	err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		return nil, err
	}

	return &TestS3Storage{
		container: minioContainer,
		client:    client,
		bucket:    bucketName,
	}, nil
}

func (ts *TestS3Storage) Terminate(ctx context.Context) error {
	if ts.container != nil {
		return ts.container.Terminate(ctx)
	}
	return nil
}

func (ts *TestS3Storage) CleanBucket(ctx context.Context) error {
	objectsCh := ts.client.ListObjects(ctx, ts.bucket, minio.ListObjectsOptions{
		Recursive: true,
	})

	for object := range objectsCh {
		if object.Err != nil {
			return object.Err
		}
		err := ts.client.RemoveObject(ctx, ts.bucket, object.Key, minio.RemoveObjectOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (ts *TestS3Storage) GetClient() *minio.Client {
	return ts.client
}

func (ts *TestS3Storage) GetBucket() string {
	return ts.bucket
}

func (ts *TestS3Storage) GetEndpoint(ctx context.Context) (string, error) {
	host, err := ts.container.Host(ctx)
	if err != nil {
		return "", err
	}

	port, err := ts.container.MappedPort(ctx, "9000")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s:%s", host, port.Port()), nil
}

func (ts *TestS3Storage) GetCredentials() (accessKey, secretKey string) {
	return "minioadmin", "minioadmin"
}
