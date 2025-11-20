package ingest

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/NBISweden/submitter/internal/models"
)

type mockClient struct {
	FilesToReturn []models.FileInfo
	Response      *http.Response
	CallIndex     int
}

func (m *mockClient) GetUsersFiles() ([]models.FileInfo, error) {
	return m.FilesToReturn, nil
}

func (m *mockClient) PostFileIngest(data []byte) (*http.Response, error) {
	return m.Response, nil
}

func (m *mockClient) PostFileAccession(payload []byte) (*http.Response, error) {
	return m.Response, nil
}

func setup(userID string, datasetFolder string) *mockClient {
	mock := &mockClient{
		FilesToReturn: []models.FileInfo{
			{InboxPath: fmt.Sprintf("/%s/%s/file1.c4gh", userID, datasetFolder), Status: "uploaded"},
			{InboxPath: fmt.Sprintf("/%s/%s/file2.c4gh", userID, datasetFolder), Status: "uploaded"},
			{InboxPath: fmt.Sprintf("/%s/PRIVATE/%s/file4.c4gh", userID, datasetFolder), Status: "uploaded"},
			{InboxPath: fmt.Sprintf("/%s/%s/file5.c4gh", userID, datasetFolder), Status: "error"},
		},
		Response: &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString("ok"))},
	}
	return mock
}

func TestIngest(t *testing.T) {
	userID := "testuser"
	datasetFolder := "DATASET_TEST"
	expectedFiles := 2
	mock := setup(userID, datasetFolder)

	t.Run("Test Ingest", func(t *testing.T) {
		userFiles, err := mock.GetUsersFiles()
		if err != nil {
			t.Error(err)
		}
		files, err := ingestFiles(mock, datasetFolder, userID, userFiles)
		if err != nil {
			t.Error(err)
		}
		if files != expectedFiles {
			t.Logf("ingested %d/%d files", files, expectedFiles)
			t.FailNow()
		}
		t.Logf("ingested %d/%d files sucessfully", files, expectedFiles)
	})
}
