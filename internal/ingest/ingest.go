package ingest

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/NBISweden/submitter/pkg/sdaclient"
	"github.com/schollz/progressbar/v3"
)

type File struct {
	InboxPath  string `json:"inboxPath"`
	FileStatus string `json:"fileStatus"`
}

func IngestFiles(sdaclient *sdaclient.Client, dryRun bool) (int, error) {

	fmt.Println("[Ingest] Waiting on response from sda api ...")
	response, err := sdaclient.GetUsersFiles()
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}

	var files []File
	if err := json.Unmarshal(body, &files); err != nil {
		return 0, err
	}

	var fileList []string
	for _, f := range files {
		if f.FileStatus != "uploaded" {
			continue
		}
		if !strings.Contains(f.InboxPath, sdaclient.DatasetFolder) {
			continue
		}
		if strings.Contains(f.InboxPath, "PRIVATE") || strings.Contains(f.InboxPath, "LANDING PAGE") {
			continue
		}
		fileList = append(fileList, f.InboxPath)
	}

	filesCount := len(fileList)
	fmt.Printf("[Ingest] Number of files to ingest: %d\n", filesCount)
	if dryRun {
		fmt.Println("[Dry-Run] Files will not be ingested")
		return filesCount, nil
	}

	bar := progressbar.Default(int64(len(fileList)), "[Ingest] Running ingestion")
	for _, path := range fileList {
		bar.Add(1)
		payload := map[string]string{
			"filepath": path,
			"user":     sdaclient.UserID,
		}
		data, _ := json.Marshal(payload)

		response, err = sdaclient.PostFileIngest(data)
		if err != nil {
			return filesCount, err
		}

		io.Copy(io.Discard, response.Body)
		response.Body.Close()
	}

	fmt.Printf("[Ingest] Starting ingest queue for %d files\n", filesCount)
	return filesCount, nil
}
