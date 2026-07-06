package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorage struct {
	client *minio.Client
	bucket string
}

func NewMinioStorage(
	endpoint,
	accessKey,
	secretKey,
	bucket string,
	useSSL bool,
) (*MinioStorage, error) {

	client, err := minio.New(endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			accessKey,
			secretKey,
			"",
		),
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
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
	}

	return &MinioStorage{
		client: client,
		bucket: bucket,
	}, nil
}

func (m *MinioStorage) UploadProductImage(
	file multipart.File,
	header *multipart.FileHeader,
	productID uint,
) (string, error) {

	defer file.Close()

	extension := filepath.Ext(header.Filename)

	filename := uuid.New().String() + extension

	objectName := fmt.Sprintf(
		"products/%d/%s",
		productID,
		filename,
	)

	_, err := m.client.PutObject(
		context.Background(),
		m.bucket,
		objectName,
		file,
		header.Size,
		minio.PutObjectOptions{
			ContentType: header.Header.Get("Content-Type"),
		},
	)

	if err != nil {
		return "", err
	}

	return objectName, nil
}
