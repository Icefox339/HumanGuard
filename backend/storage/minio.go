package storage

import (
    "context"
    "io"
    "log"

    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOClient struct {
    client *minio.Client
    bucket string
}

func NewMinIOClient(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinIOClient, error) {
    client, err := minio.New(endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
        Secure: useSSL,
    })
    if err != nil {
        return nil, err
    }

    ctx := context.Background()
    exists, err := client.BucketExists(ctx, bucket)
    if err != nil {
        return nil, err
    }
    if !exists {
        if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
            return nil, err
        }
        log.Printf("Created bucket: %s", bucket)
    }

    return &MinIOClient{client: client, bucket: bucket}, nil
}

func (m *MinIOClient) Save(path string, reader io.Reader) (int64, error) {
    info, err := m.client.PutObject(context.Background(), m.bucket, path, reader, -1, minio.PutObjectOptions{})
    if err != nil {
        return 0, err
    }
    return info.Size, nil
}

func (m *MinIOClient) Get(path string) (io.ReadCloser, error) {
    return m.client.GetObject(context.Background(), m.bucket, path, minio.GetObjectOptions{})
}

func (m *MinIOClient) Delete(path string) error {
    return m.client.RemoveObject(context.Background(), m.bucket, path, minio.RemoveObjectOptions{})
}