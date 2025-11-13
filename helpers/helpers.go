package helpers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/NBISweden/submitter/internal/client"
)

type File struct {
	InboxPath  string `json:"inboxPath"`
	FileStatus string `json:"fileStatus"`
}

// Move this to client.go package?
func WaitForAccession(api *client.Client, target int, interval time.Duration, timeout time.Duration) ([]string, error) {
	deadline := time.Now().Add(timeout)
	for {
		paths, err := getVerifiedFilePaths(api)
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

func getVerifiedFilePaths(client *client.Client) ([]string, error) {
	response, err := client.GetUsersFiles()
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
			strings.Contains(f.InboxPath, client.DatasetFolder) &&
			!strings.Contains(f.InboxPath, "PRIVATE") {
			paths = append(paths, f.InboxPath)
		}
	}
	return paths, nil
}

func GetFileIDsPath(dataDirectory string, datasetFolder string) string {
	return fmt.Sprintf("%s/%s-fileIDs.txt", dataDirectory, datasetFolder)
}

func GetStableIDsPath(dataDirectory string, datasetFolder string) string {
	return fmt.Sprintf("%s/%s-stableIDs.txt", dataDirectory, datasetFolder)
}
