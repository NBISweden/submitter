package client

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Client struct {
	accessToken   string
	apiHost       string
	userID        string
	datasetFolder string
	datasetID     string
	httpClient    *http.Client
}

type File struct {
	InboxPath  string `json:"inboxPath"`
	FileStatus string `json:"fileStatus"`
}

func New(configPath string) (*Client, error) {
	conf, err := NewConfig(configPath)
	if err != nil {
	}
	httpClient := http.DefaultClient
	if conf.ssl {
		caCert, err := os.ReadFile(conf.sslCaCert)
		if err != nil {
			return nil, fmt.Errorf("init config: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
			return nil, fmt.Errorf("read CA cert %q: %w", conf.sslCaCert, err)
		}

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		}
		httpClient = &http.Client{Transport: tr}
	}

	client := &Client{
		accessToken:   conf.accessToken,
		apiHost:       conf.apiHost,
		userID:        conf.userID,
		datasetFolder: conf.datasetFolder,
		datasetID:     conf.datasetID,
		httpClient:    httpClient,
	}

	return client, nil
}

func (c *Client) GetUsersFilesWithPrefix() (*http.Response, error) {
	basePath := fmt.Sprintf("users/%s/files", c.userID)

	u, err := url.Parse(basePath)
	if err != nil {
		return nil, fmt.Errorf("unable to parse base path: %w", err)
	}

	q := u.Query()
	q.Set("path_prefix", c.datasetFolder)
	u.RawQuery = q.Encode()

	return c.doRequest("GET", u.String(), nil)
}

func (c *Client) GetUsersFiles() (*http.Response, error) {
	return c.doRequest("GET", fmt.Sprintf("users/%s/files", c.userID), nil)
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
	url := fmt.Sprintf("%s/%s", c.apiHost, path)
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	slog.Info("[client] calling", "method", method, "url", url)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	slog.Info("[client] response", "status", resp.Status)
	return resp, nil
}

func (c *Client) WaitForAccession(target int, interval time.Duration, timeout time.Duration) ([]string, error) {
	deadline := time.Now().Add(timeout)
	for {
		paths, err := c.getVerifiedFilePaths()
		if err != nil {
			return nil, err
		}

		if len(paths) >= target {
			return paths, nil
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout reached, only got %d/%d files", len(paths), target)
		}
		slog.Info(fmt.Sprintf("[accession] found %d/%d files - waiting: internal: %s timeout: %s", len(paths), target, interval, timeout))
		time.Sleep(interval)
	}
}

func (c *Client) getVerifiedFilePaths() ([]string, error) {
	response, err := c.GetUsersFiles()
	if err != nil {
		return nil, err
	}
	defer response.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("[accession] failed to read response body %w", err)
	}

	var files []File
	if err := json.Unmarshal(body, &files); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user files: %w", err)
	}

	var paths []string
	for _, f := range files {
		if f.FileStatus == "verified" &&
			strings.Contains(f.InboxPath, c.datasetFolder) &&
			!strings.Contains(f.InboxPath, "PRIVATE") {
			paths = append(paths, f.InboxPath)
		}
	}
	return paths, nil
}
