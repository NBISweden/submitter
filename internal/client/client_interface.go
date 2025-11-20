package client

import (
	"net/http"

	"github.com/NBISweden/submitter/internal/models"
)

type APIClient interface {
	GetUsersFiles() ([]models.FileInfo, error)
	PostFileIngest([]byte) (*http.Response, error)
	PostFileAccession(payload []byte) (*http.Response, error)
}
