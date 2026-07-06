package storage

import (
	"context"
	"mime/multipart"

	"github.com/minio/minio-go/v7"
)

type Storage interface {
	UploadProductImage(
		file multipart.File,
		header *multipart.FileHeader,
		productID uint,
	) (string, error)

	DeleteObject(objectName string) error
}

func (m *MinioStorage) DeleteObject(objectName string) error {
	return m.client.RemoveObject(
		context.Background(),
		m.bucket,
		objectName,
		minio.RemoveObjectOptions{},
	)
}
