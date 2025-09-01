package ingest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type File struct {
	InboxPath  string `json:"inboxPath"`
	FileStatus string `json:"fileStatus"`
}

func IngestFiles(accessToken, apiHost, user, datasetFolder string, dryRun bool) error {
	url := fmt.Sprintf("%s/users/%s/files", apiHost, user)
	fmt.Println("Calling:", url)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	response, err := client.Do(request)
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
		if !strings.Contains(f.InboxPath, datasetFolder) {
			continue
		}
		if strings.Contains(f.InboxPath, "PRIVATE") || strings.Contains(f.InboxPath, "LANDING PAGE") {
			continue
		}
		fileList = append(fileList, f.InboxPath)
	}

	filesCount := len(fileList)
	fmt.Printf("Number of files to ingest: %d\n", filesCount)
	if dryRun {
		fmt.Println("Dry run, not ingesting files")
		return nil
	}

	for _, path := range fileList {
		payload := map[string]string{
			"filepath": path,
			"user":     user,
		}
		data, _ := json.Marshal(payload)

		url := apiHost + "/file/ingest"
		fmt.Println("Calling:", url)
		request, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
		if err != nil {
			return err
		}

		request.Header.Set("Authorization", "Bearer "+accessToken)
		request.Header.Set("Content-Type", "application/json")

		response, err := client.Do(request)
		if err != nil {
			return err
		}
		io.Copy(io.Discard, response.Body)
		response.Body.Close()
	}

	fmt.Println("Messages have been sent to Ingest queue")
	return nil
}
