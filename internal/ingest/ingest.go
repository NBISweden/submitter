package ingest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/NBISweden/submitter/pkg/sdaclient"
	"github.com/schollz/progressbar/v3"
)

type File struct {
	InboxPath  string `json:"inboxPath"`
	FileStatus string `json:"fileStatus"`
}

func IngestFiles(sdaclient *sdaclient.Client, dryRun bool) (int, error) {

	fmt.Println("[ingest] waiting on response from sda api ...")
	response, err := sdaclient.GetUsersFiles()
	if err != nil {
		return 0, err
	}
	defer response.Body.Close() //nolint:errcheck

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
	fmt.Printf("[ingest] number of files to ingest: %d\n", filesCount)
	if dryRun {
		fmt.Println("[dry-run] files will not be ingested")
		return filesCount, nil
	}

	var resendPayloads []map[string]string
	var nonOKResponds []int
	var okResponds []int

	bar := progressbar.Default(int64(len(fileList)), "[ingest] running ingestion")
	for _, path := range fileList {
		_ = bar.Add(1)
		payload := map[string]string{
			"filepath": path,
			"user":     sdaclient.UserID,
		}
		data, _ := json.Marshal(payload)

		response, err = sdaclient.PostFileIngest(data)
		if err != nil {
			return filesCount, err
		}

		if response.StatusCode != http.StatusOK {
			nonOKResponds = append(nonOKResponds, response.StatusCode)
			resendPayloads = append(resendPayloads, payload)
		}

		if response.StatusCode == http.StatusOK {
			okResponds = append(okResponds, response.StatusCode)
		}

		io.Copy(io.Discard, response.Body) //nolint:errcheck
		response.Body.Close() //nolint:errcheck
	}

	if len(resendPayloads) != 0 {
		fmt.Printf("[ingest] warning! Found %d non-ok responds from SDA api\n", len(resendPayloads))
		countResponds := make(map[int]int)
		for _, code := range nonOKResponds {
			countResponds[code]++
		}

		for code, count := range countResponds {
			fmt.Printf("[ingest] found: %d with status: %d\n", count, code)
		}
	}

	fmt.Printf("[ingest] started ingest queue for %d files\n", len(okResponds))
	return filesCount, nil
}
