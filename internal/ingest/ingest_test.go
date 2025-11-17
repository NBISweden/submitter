package ingest

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/NBISweden/submitter/internal/database"
)

type mockClient struct {
	FilesToReturn []*database.SubmissionFileInfo
	Response      *http.Response
	CallIndex     int
}

func (m *mockClient) GetUsersFiles() ([]*database.SubmissionFileInfo, error) {
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
		FilesToReturn: []*database.SubmissionFileInfo{
			{InboxPath: fmt.Sprintf("/%s/%s/file1.c4gh", userID, datasetFolder), Status: "uploaded"},
			{InboxPath: fmt.Sprintf("/%s/%s/file2.c4gh", userID, datasetFolder), Status: "uploaded"},
			{InboxPath: fmt.Sprintf("/someuser/%s/file3.c4gh", datasetFolder), Status: "uploaded"},
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
		files, err := IngestFiles(mock, userID, datasetFolder)
		if err != nil {
			t.Fail()
		}
		if files != expectedFiles {
			t.Logf("ingested %d/%d files", files, expectedFiles)
			t.FailNow()
		}
		t.Logf("ingested %d/%d files sucessfully", files, expectedFiles)
	})
}
