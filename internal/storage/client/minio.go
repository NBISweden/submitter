package storage

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/NBISweden/sda-bpctl/internal/storage"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage struct {
	client *minio.Client
}

func NewMinio(endpoint, accessKey, secretKey, sslCaCert string) (*Storage, error) {
	var caCertPool *x509.CertPool
	if sslCaCert != "" {
		caCert, err := os.ReadFile(sslCaCert)
		if err != nil {
			return nil, fmt.Errorf("init config: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
			return nil, fmt.Errorf("read CA cert %q: %w", sslCaCert, err)
		}
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: caCertPool,
		},
		TLSHandshakeTimeout: 15 * time.Second,
	}

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:    true,
		Transport: tr,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create s3client on '%s': %v", endpoint, err)
	}

	storage := &Storage{
		client: minioClient,
	}
	return storage, nil
}

func (s *Storage) ListObjects(bucket, prefix string) ([]storage.ObjectInfo, error) {
	var objects []storage.ObjectInfo
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	objectCh := s.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})
	for object := range objectCh {
		if object.Err != nil {
			return objects, fmt.Errorf("err when listing object %v", object.Err)
		}
		objects = append(objects, storage.ObjectInfo{Key: object.Key, Size: object.Size})
	}
	return objects, nil
}

func (s *Storage) GetObject(bucketName, objectName string) (io.ReadCloser, error) {
	return s.client.GetObject(context.Background(), bucketName, objectName, minio.GetObjectOptions{})
}

func (s *Storage) PutObject(bucketName, objectName string, reader io.Reader, size int64) error {
	_, err := s.client.PutObject(context.Background(), bucketName, objectName, reader, size, minio.PutObjectOptions{})
	return err
}
