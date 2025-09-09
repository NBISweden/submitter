package ingest

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/NBISweden/submitter/pkg/sdaclient"
)

type File struct {
	InboxPath  string `json:"inboxPath"`
	FileStatus string `json:"fileStatus"`
}

func IngestFiles(sdaclient *sdaclient.Client, dryRun bool) error {

	response, err := sdaclient.GetUsersFiles()
	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var files []File
	if err := json.Unmarshal(body, &files); err != nil {
		return err
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
	fmt.Printf("[Ingest] Number of files to ingest : %d\n", filesCount)
	if dryRun {
		fmt.Println("[Dry-Run] Files will not be ingested")
		// Don't ask me why, but If I don't have this print the print above will not show
		fmt.Println()
		return nil
	}

	for _, path := range fileList {
		payload := map[string]string{
			"filepath": path,
			"user":     sdaclient.UserID,
		}
		data, _ := json.Marshal(payload)

		response, err = sdaclient.PostFileIngest(data)
		if err != nil {
			return err
		}

		io.Copy(io.Discard, response.Body)
		response.Body.Close()
	}

	fmt.Println("Messages have been sent to Ingest queue")
	return nil
}
