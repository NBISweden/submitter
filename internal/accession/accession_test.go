package accession

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/NBISweden/submitter/internal/database"
)

type mockClient struct {
	FilesToReturn []*database.SubmissionFileInfo
	Response      *http.Response
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

func newMockClient(userID string, datasetFolder string) *mockClient {
	mock := &mockClient{
		FilesToReturn: []*database.SubmissionFileInfo{
			{InboxPath: fmt.Sprintf("/%s/%s/file1.c4gh", userID, datasetFolder), Status: "verified"},
			{InboxPath: fmt.Sprintf("/%s/%s/file2.c4gh", userID, datasetFolder), Status: "verified"},
			{InboxPath: fmt.Sprintf("/%s/%s/file3.c4gh", userID, datasetFolder), Status: "error"},
		},
		Response: &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString("ok"))},
	}
	return mock
}

func TestAccession(t *testing.T) {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	workingDirectory := filepath.Dir(ex)
	userID := "testuser"
	datasetFolder := "DATASET_TEST"
	mock := newMockClient(userID, datasetFolder)

	t.Run("Test Accession", func(t *testing.T) {
		accessionCmd.Flag("data-directory").Value.Set(workingDirectory)
		err := CreateAccessionIDs(mock, datasetFolder, userID)
		if err != nil {
			t.Error(err)
		}
	})

}
