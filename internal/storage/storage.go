package storage

import (
	"io"
)

type ObjectInfo struct {
	Key  string
	Size int64
}

type StorageClient interface {
	ListObjects(bucket, prefix string) ([]ObjectInfo, error)
	GetObject(bucketName, objectName string) (io.ReadCloser, error)
	PutObject(bucketName, objectName string, reader io.Reader, size int64) error
	RemoveObject(bucketName, objectName string, reader io.Reader) error
}
