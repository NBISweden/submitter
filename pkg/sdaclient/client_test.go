package sdaclient

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGet_Files(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/testuser/files" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer mockServer.Close()

	mockClient := Client{
		AccessToken:   "test-token",
		APIHost:       mockServer.URL,
		UserID:        "testuser",
		DatasetFolder: "test-folder",
		HTTPClient:    http.DefaultClient,
	}

	response, err := mockClient.GetUsersFiles()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected http response %d, got %d", http.StatusOK, response.StatusCode)
	}

}
