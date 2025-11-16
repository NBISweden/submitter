package client

import (
	"net/http"

	"github.com/NBISweden/submitter/internal/database"
)

type APIClient interface {
	GetUsersFiles() ([]*database.SubmissionFileInfo, error)
	PostFileIngest([]byte) (*http.Response, error)
	PostFileAccession(payload []byte) (*http.Response, error)
}
