package sdaclient

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"

	"github.com/schollz/progressbar/v3"
)

type Client struct {
	AccessToken   string
	APIHost       string
	UserID        string
	DatasetFolder string
	DatasetID     string
	HTTPClient    *http.Client
}

func NewClient(token string, apiHost string, userID string, datasetFolder string, datasetID string) *Client {
	return &Client{
		AccessToken:   token,
		APIHost:       apiHost,
		UserID:        userID,
		DatasetFolder: datasetFolder,
		DatasetID:     datasetID,
		HTTPClient:    http.DefaultClient,
	}
}

func (c *Client) GetUsersFilesWithPrefix() (*http.Response, error) {
	basePath := fmt.Sprintf("users/%s/files", c.UserID)

	u, err := url.Parse(basePath)
	if err != nil {
		return nil, fmt.Errorf("unable to parse base path: %w", err)
	}

	q := u.Query()
	q.Set("path_prefix", c.DatasetFolder)
	u.RawQuery = q.Encode()

	return c.doRequest("GET", u.String(), nil)
}

func (c *Client) GetUsersFiles() (*http.Response, error) {
	return c.doRequest("GET", fmt.Sprintf("users/%s/files", c.UserID), nil)
}

func (c *Client) PostFileIngest(payload []byte) (*http.Response, error) {
	return c.doRequest("POST", "file/ingest", payload)
}

func (c *Client) PostFileAccession(payload []byte) (*http.Response, error) {
	return c.doRequest("POST", "file/accession", payload)
}

func (c *Client) PostDatasetCreate(payload []byte) (*http.Response, error) {
	return c.doRequest("POST", "dataset/create", payload)
}

func (c *Client) doRequest(method, path string, body []byte) (*http.Response, error) {
	bar := progressbar.Default(-1, "[Client] Waiting for SDA API")
	url := fmt.Sprintf("%s/%s", c.APIHost, path)
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating %s request to %s: %w", method, url, err)
	}
	bar.Add(1)

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing %s request to %s: %w", method, url, err)
	}

	return resp, nil
}
