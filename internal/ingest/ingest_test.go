package ingest

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NBISweden/submitter/pkg/sdaclient"
)

func TestIngestFiles_HappyPath(t *testing.T) {
	// Mock files response (mix of uploaded and non-uploaded)
	files := []File{
		{InboxPath: "dataset1/file1.txt", FileStatus: "uploaded"},
		{InboxPath: "dataset1/file2.txt", FileStatus: "processing"},
		{InboxPath: "dataset1/PRIVATE/file3.txt", FileStatus: "uploaded"},
		{InboxPath: "dataset1/file4.txt", FileStatus: "uploaded"},
		{InboxPath: "otherDataset/file5.txt", FileStatus: "uploaded"},
	}
	filesJSON, _ := json.Marshal(files)

	var postRequests []string

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/users/") && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(filesJSON)
			return
		}
		if r.URL.Path == "/file/ingest" && r.Method == http.MethodPost {
			body, _ := io.ReadAll(r.Body)
			postRequests = append(postRequests, string(body))
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer mockServer.Close()

	accessToken := "test-token"
	user := "testuser"
	datasetFolder := "dataset1"
	dryRun := false
	sdaclient := &sdaclient.Client{
		AccessToken:   accessToken,
		APIHost:       mockServer.URL,
		UserID:        user,
		DatasetFolder: datasetFolder,
		HTTPClient:    http.DefaultClient,
	}

	err := IngestFiles(sdaclient, dryRun)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify POST requests count (should be 2: file1.txt and file4.txt)
	expectedPosts := 2
	if len(postRequests) != expectedPosts {
		t.Fatalf("expected %d POST requests, got %d", expectedPosts, len(postRequests))
	}

	// Check that posted file paths match expected
	for _, body := range postRequests {
		if !strings.Contains(body, "file1.txt") && !strings.Contains(body, "file4.txt") {
			t.Errorf("unexpected file in POST body: %s", body)
		}
	}
}
