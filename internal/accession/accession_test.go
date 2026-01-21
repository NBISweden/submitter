package accession

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/NBISweden/sda-bpctl/internal/models"
)

type mockClient struct {
	FilesToReturn []models.FileInfo
	Response      *http.Response
}

func (m *mockClient) GetUsersFilesWithPrefix() ([]models.FileInfo, error) {
	return m.FilesToReturn, nil
}

func (m *mockClient) GetUsersFiles() ([]models.FileInfo, error) {
	return m.FilesToReturn, nil
}

func (m *mockClient) PostFileIngest(data []byte) ([]byte, error) {
	response, err := json.Marshal(m.Response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (m *mockClient) PostFileAccession(payload []byte) ([]byte, error) {
	response, err := json.Marshal(m.Response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func newMockClient(userID string, datasetFolder string) *mockClient {
	mock := &mockClient{
		FilesToReturn: []models.FileInfo{
			// Data is mocked so we expect 2 files to be included in the accession (file1.c4gh and file2.c4gh)
			{InboxPath: fmt.Sprintf("/%s/%s/file1.c4gh", userID, datasetFolder), Status: "verified"},
			{InboxPath: fmt.Sprintf("/%s/%s/file2.c4gh", userID, datasetFolder), Status: "verified"},
			{InboxPath: fmt.Sprintf("/%s/%s/file3.c4gh", userID, datasetFolder), Status: "error"},
			{InboxPath: fmt.Sprintf("/%s/%s/file4.c4gh", userID, "DATASET_OTHER"), Status: "verified"},
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
	datasetFolder = "DATASET_TEST"
	expectedPaths := 2
	mock := newMockClient(userID, datasetFolder)

	t.Run("Test Accession", func(t *testing.T) {

		accessionCmd.Flag("data-directory").Value.Set(workingDirectory)
		files, err := mock.GetUsersFiles()
		if err != nil {
			t.Error(err)
		}

		paths := getPathsForAccessionIDs(files)
		nrPaths := len(paths)
		if nrPaths != expectedPaths {
			t.Logf("recieved %d/%d paths for accessionIDs", nrPaths, expectedPaths)
			t.Fail()
		}

		accessionIDs, err := postAccessionIDs(mock, paths, userID)
		nrAccessionIDs := len(accessionIDs)
		if nrAccessionIDs != expectedPaths {
			t.Logf("recieved %d/%d accessionIDs", nrAccessionIDs, expectedPaths)
			t.Fail()
		}
	})

}
