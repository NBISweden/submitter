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
	"strings"
	"time"

	"github.com/NBISweden/submitter/internal/database"
	"github.com/cenkalti/backoff/v4"
)

type Client struct {
	accessToken    string
	apiHost        string
	userID         string
	datasetFolder  string
	datasetID      string
	httpClient     *http.Client
	postgresClient *database.PostgresDb
}

func New(configPath string) (*Client, error) {
	conf, err := NewConfig(configPath)
	if err != nil {
		return nil, err
	}

	postgresClient, err := database.New(configPath)
	if err != nil {
		return nil, err
	}

	httpClient := http.DefaultClient
	if conf.ssl {
		caCert, err := os.ReadFile(conf.caCert)
		if err != nil {
			return nil, fmt.Errorf("init config: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
			return nil, fmt.Errorf("read CA cert %q: %w", conf.caCert, err)
		}

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		}
		httpClient = &http.Client{Transport: tr}
	}

	client := &Client{
		accessToken:    conf.accessToken,
		apiHost:        conf.apiHost,
		userID:         conf.userID,
		datasetFolder:  conf.datasetFolder,
		datasetID:      conf.datasetID,
		httpClient:     httpClient,
		postgresClient: postgresClient,
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

func (c *Client) GetUsersFiles() ([]*database.SubmissionFileInfo, error) {
	var files []*database.SubmissionFileInfo
	err := backoff.Retry(func() error {
		var err error
		files, err = c.postgresClient.GetUserFiles(c.userID, c.datasetFolder, true)
		return err
	}, backoff.NewExponentialBackOff())
	if err != nil {
		return nil, err
	}
	return files, nil
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
	var req *http.Request
	var resp *http.Response
	err := backoff.Retry(func() error {
		var err error
		url := fmt.Sprintf("%s/%s", c.apiHost, path)
		req, err = http.NewRequest(method, url, bytes.NewReader(body))
		if err != nil {
			return err
		}

		req.Header.Set("Authorization", "Bearer "+c.accessToken)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		slog.Info("request", "method", method, "url", url)
		resp, err = c.httpClient.Do(req)
		if err != nil {
			return err
		}
		return nil
	}, backoff.NewExponentialBackOff())

	if err != nil {
		return resp, err
	}
	slog.Info("response", "status", resp.Status)
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
	files, err := c.GetUsersFiles()
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, f := range files {
		if f.Status == "verified" &&
			strings.Contains(f.InboxPath, c.datasetFolder) &&
			!strings.Contains(f.InboxPath, "PRIVATE") {
			paths = append(paths, f.InboxPath)
		}
	}
	return paths, nil
}
