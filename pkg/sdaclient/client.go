package sdaclient

import (
	"bytes"
	"fmt"
	"net/http"
)

type Client struct {
	AccessToken   string
	APIHost       string
	UserID        string
	DatasetFolder string
	HTTPClient    *http.Client
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

func (c *Client) doRequest(method, path string, body []byte) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s", c.APIHost, path)
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating %s request to %s: %w", method, url, err)
	}

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
