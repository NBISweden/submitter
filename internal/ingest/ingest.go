package ingest

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/NBISweden/submitter/cmd"
	"github.com/NBISweden/submitter/internal/client"
	"github.com/NBISweden/submitter/internal/config"
	"github.com/NBISweden/submitter/internal/database"
	"github.com/spf13/cobra"
)

var dryRun bool
var configPath string

type APIClient interface {
	GetUsersFiles() ([]*database.SubmissionFileInfo, error)
	PostFileIngest([]byte) (*http.Response, error)
}

var ingestCmd = &cobra.Command{
	Use:   "ingest [flags]",
	Short: "Trigger ingestion",
	Long:  "Trigger ingestion",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := config.NewConfig(configPath)
		if err != nil {
			return err
		}
		api, err := client.New(configPath)
		if err != nil {
			return err
		}
		_, err = IngestFiles(api, conf.DatasetFolder, conf.UserID)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	cmd.AddCommand(ingestCmd)
	ingestCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Toggles dry-run mode. Dry run will not run any state changing API calls")
	ingestCmd.Flags().StringVar(&configPath, "config", "config.yaml", "Path to configuration file")
}

func IngestFiles(api APIClient, datasetFolder string, userID string) (int, error) {
	files, err := api.GetUsersFiles()
	if err != nil {
		return 0, err
	}

	var fileList []string
	for _, f := range files {
		if f.Status != "uploaded" {
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
	slog.Info("[ingest] number of files to ingest", "filesCount", filesCount)
	if dryRun {
		slog.Info("[ingest] dry-run enabled. No files will be ingested")
		return filesCount, nil
	}

	var resendPayloads []map[string]string
	var nonOKResponds []int
	var okResponds []int

	for _, path := range fileList {
		payload := map[string]string{
			"filepath": path,
			"user":     userID,
		}
		data, _ := json.Marshal(payload)

		response, err := api.PostFileIngest(data)
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
		response.Body.Close()              //nolint:errcheck
	}

	if len(resendPayloads) != 0 {
		slog.Warn("[ingest] found non-ok responds from SDA API", "non-oks", len(resendPayloads))
		countResponds := make(map[int]int)
		for _, code := range nonOKResponds {
			countResponds[code]++
		}

		for code, count := range countResponds {
			slog.Info("[ingest] non-ok responds", "count", count, "code", code)
		}
	}

	slog.Info(fmt.Sprintf("[ingest] starting ingestion for %d/%d successful responses", len(okResponds), filesCount))
	return filesCount, nil
}
