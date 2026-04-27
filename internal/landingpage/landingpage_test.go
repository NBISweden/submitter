package landingpage

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/NBISweden/sda-bpctl/internal/storage"
)

type mockStorage struct {
	FilesToReturn []storage.ObjectInfo
}

func NewMockInbox() *mockStorage {
	return &mockStorage{
		FilesToReturn: []storage.ObjectInfo{{Key: "bucket/file1.jpg.c4gh", Size: 0}, {Key: "bucket/file2.jpg.c4gh", Size: 0}},
	}
}

func (m *mockStorage) ListObjects(bucket, prefix string) ([]storage.ObjectInfo, error) {
	return m.FilesToReturn, nil
}

func (m *mockStorage) GetObject(bucketName, objectName string) (io.ReadCloser, error) {
	objectLocation := fmt.Sprintf("%s/%s", bucketName, objectName)
	for _, object := range m.FilesToReturn {
		if object.Key == objectLocation {
			return io.NopCloser(bytes.NewBuffer([]byte("somebytes"))), nil
		}
	}
	return io.NopCloser(bytes.NewBuffer([]byte("somebytes"))), nil
}

func (m *mockStorage) PutObject(bucketName, objectName string, reader io.Reader, size int64) error {
	return nil
}
func (m *mockStorage) RemoveObject(bucketName, objectName string, reader io.Reader) error {
	return nil
}

func TestIngest(t *testing.T) {
	t.Log("test landingpage")

}
