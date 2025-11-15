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
	Responses     []*http.Response
	CallIndex     int
}

func setup(user string, datasetFolder string) *mockClient {
	mock := &mockClient{
		FilesToReturn: []*database.SubmissionFileInfo{
			{InboxPath: fmt.Sprintf("/%s/%s/file1.c4gh", user, datasetFolder), Status: "uploaded"},
			{InboxPath: fmt.Sprintf("/%s/%s/file2.c4gh", user, datasetFolder), Status: "uploaded"},
			{InboxPath: fmt.Sprintf("/someuser/%s/file3.c4gh", datasetFolder), Status: "uploaded"},
			{InboxPath: fmt.Sprintf("/%s/PRIVATE/%s/file4.c4gh", user, datasetFolder), Status: "uploaded"},
			{InboxPath: fmt.Sprintf("/%s/%s/file5.c4gh", user, datasetFolder), Status: "error"},
		},
		Responses: []*http.Response{
			{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString("ok"))},
			{StatusCode: http.StatusInternalServerError, Body: io.NopCloser(bytes.NewBufferString("fail"))},
		},
	}
	return mock
}

func TestIngest(t *testing.T) {
	user := "testuser"
	datasetFolder := "DATASET_TEST"
	expectedFiles := 2
	mock := setup(user, datasetFolder)

	t.Run("Test GetUsersFiles", func(t *testing.T) {
		files, err := IngestFiles(mock, user, datasetFolder)
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

func (m *mockClient) GetUsersFiles() ([]*database.SubmissionFileInfo, error) {
	return m.FilesToReturn, nil
}

func (m *mockClient) PostFileIngest(data []byte) (*http.Response, error) {
	if m.CallIndex >= len(m.Responses) {
		return nil, nil
	}
	resp := m.Responses[m.CallIndex]
	m.CallIndex++
	return resp, nil
}
