package client

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/NBISweden/submitter/internal/config"
)

type Client struct {
	AccessToken   string
	APIHost       string
	UserID        string
	DatasetFolder string
	DatasetID     string
	HTTPClient    *http.Client
}

func NewClient(conf config.Config) *Client {
	var httpClient *http.Client
	httpClient = http.DefaultClient
	if conf.UseTLS {
		caCert, err := os.ReadFile(conf.SSLCACert)
		if err != nil {
			slog.Error("error", "err", err)
			return nil
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		}
		httpClient = &http.Client{Transport: tr}
	}

	return &Client{
		AccessToken:   conf.AccessToken,
		APIHost:       conf.APIHost,
		UserID:        conf.UserID,
		DatasetFolder: conf.DatasetFolder,
		DatasetID:     conf.DatasetID,
		HTTPClient:    httpClient,
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
	url := fmt.Sprintf("%s/%s", c.APIHost, path)
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	slog.Info("calling", "method", method, "url", url)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	slog.Info("Response", "status", resp.Status)
	return resp, nil
}
