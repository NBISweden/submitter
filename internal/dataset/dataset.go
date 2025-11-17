package dataset

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"slices"

	"github.com/NBISweden/submitter/cmd"
	"github.com/NBISweden/submitter/helpers"
	"github.com/NBISweden/submitter/internal/client"
	"github.com/NBISweden/submitter/internal/config"
	"github.com/spf13/cobra"
)

var dryRun bool
var configPath string
var dataDirectory string

var datasetCmd = &cobra.Command{
	Use:   "dataset [flags]",
	Short: "Trigger dataset creation",
	Long:  "Trigger dataset creation",
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
		err = CreateDataset(api, conf.DatasetFolder, conf.DatasetID, conf.UserID)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	cmd.AddCommand(datasetCmd)
	datasetCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Toggles dry-run mode. Dry run will not run any state changing API calls")
	datasetCmd.Flags().StringVar(&configPath, "config", "config.yaml", "Path to configuration file")
	datasetCmd.Flags().StringVar(&dataDirectory, "data-directory", "data", "Path to directory to write / read intermediate files for stableIDs and fileIDs")
}

var ErrFileAlreadyExists = errors.New("file already exists")

type Payload struct {
	AccessionIDs []string `json:"accession_ids"`
	DatasetID    string   `json:"dataset_id"`
	User         string   `json:"user"`
}

type UserFiles struct {
	AccessionID string `json:"accessionID"`
	InboxPath   string `json:"inboxPath"`
}

func CreateDataset(api *client.Client, datasetFolder string, datasetID string, userID string) error {
	slog.Info("starting dataset")
	if !dryRun {
		err := createStableIDsFile(api, datasetFolder)
		if err != nil {
			slog.Error("failed to create file with stable ids")
		}
	}

	var fileIDsList []string
	filePath := helpers.GetFileIDsPath(dataDirectory, datasetFolder)
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close() //nolint:errcheck

	slog.Info("reading", "filePath", filePath)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fileIDsList = append(fileIDsList, scanner.Text())
	}
	slog.Info("nr of files included in dataset", "nr_files", (len(fileIDsList)))
	if dryRun {
		slog.Info("dry-run enabled, no dataset will be created")
		return nil
	}

	if len(fileIDsList) > 100 {
		err := sendInChunks(fileIDsList, api, datasetID, userID)
		if err != nil {
			return err
		}
	}

	if len(fileIDsList) <= 100 {
		payload := Payload{
			AccessionIDs: fileIDsList,
			DatasetID:    datasetID,
			User:         userID,
		}
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		response, err := api.PostDatasetCreate(jsonData)
		if err != nil {
			// Se comment bellow in sendInChunks() why this might be needed
			if errors.Is(err, io.ErrUnexpectedEOF) {
			} else {
				return err
			}
		}
		if response.StatusCode != http.StatusOK {
			slog.Warn("got non-ok response", "status_code", response.StatusCode)
		}
		defer response.Body.Close() //nolint:errcheck
	}

	slog.Info("creation of dataset completed!")
	return nil
}

func sendInChunks(fileIDsList []string, api *client.Client, datasetID string, userID string) error {
	slog.Info("more than 100 entries, sending in chunks of 100")
	chunks := slices.Chunk(fileIDsList, 100)
	allChunks := slices.Collect(chunks)
	for _, chunk := range allChunks {
		payload := Payload{
			AccessionIDs: chunk,
			DatasetID:    datasetID,
			User:         userID,
		}
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		response, err := api.PostDatasetCreate(jsonData)
		/*
			As of 2025-09-17 we can get EOF responses when sending the accession id request to the sda api, however the request will still have been processed on the server side, but we won't get a response back since the TCP connection will be terminated.
		*/
		if err != nil {
			if errors.Is(err, io.ErrUnexpectedEOF) {
				continue
			}
			return err
		}
		if response.StatusCode != http.StatusOK {
			slog.Warn("got non-ok response", "status_code", response.StatusCode)
		}
		defer response.Body.Close() //nolint:errcheck
	}
	return nil
}

func createStableIDsFile(api *client.Client, datasetFolder string) error {
	filePath := helpers.GetStableIDsPath(dataDirectory, datasetFolder)
	if _, err := os.Stat(filePath); err == nil {
		return ErrFileAlreadyExists
	} else if !os.IsNotExist(err) {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close() //nolint:errcheck

	r, err := api.GetUsersFilesWithPrefix()
	if err != nil {
		return err
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	var stableIDs []UserFiles
	if err := json.Unmarshal(body, &stableIDs); err != nil {
		return err
	}

	for _, f := range stableIDs {
		fmt.Fprintf(file, "%s %s\n", f.AccessionID, f.InboxPath) //nolint:errcheck
	}

	slog.Info("created file with stable ids", "filePath", filePath)
	return nil
}
