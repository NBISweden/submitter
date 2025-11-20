package job

import (
	"log/slog"
	"time"

	"github.com/NBISweden/submitter/cmd"
	"github.com/NBISweden/submitter/internal/accession"
	"github.com/NBISweden/submitter/internal/client"
	"github.com/NBISweden/submitter/internal/config"
	"github.com/NBISweden/submitter/internal/database"
	"github.com/NBISweden/submitter/internal/dataset"
	"github.com/NBISweden/submitter/internal/ingest"
	"github.com/spf13/cobra"
)

var configPath string

var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "Runs all dataset submission steps as a 'job'",
	Long:  "Runs all dataset submission steps as a 'job' (ingestion, accession, dataset)",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := runJob()
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	cmd.AddCommand(jobCmd)
	jobCmd.Flags().StringVar(&configPath, "config", "config.yaml", "Path to configuration file")
}

func runJob() error {
	globalConf, err := config.NewConfig(configPath)
	if err != nil {
		return err
	}

	pollRate := time.Minute * time.Duration(globalConf.PollRate)
	timeout := time.Minute * time.Duration(globalConf.Timeout)
	datasetFolder := globalConf.DatasetFolder
	datasetID := globalConf.DatasetID
	userID := globalConf.UserID

	slog.Info("dispatching job", "dataset_folder", datasetFolder, "dataset_id", datasetID, "userID", userID)

	api, err := client.New(configPath)
	if err != nil {
		return err
	}

	db, err := database.New(configPath)
	if err != nil {
		return err
	}

	filesCount, err := ingest.Run(api, *db, datasetFolder, userID)
	if err != nil {
		return err
	}
	_, err = api.WaitForAccession(filesCount, pollRate, timeout)
	if err != nil {
		return err
	}

	accessionIDs, err := accession.Run(api, *db, datasetFolder, userID)
	if err != nil {
		return err
	}

	// We give some time for the SDA backend to process our accession ids. During test-runs it's been fine with 10 minutes
	waitTime := 10 * time.Minute
	slog.Info("waiting before sending dataset creation request", "delay", waitTime)
	time.Sleep(waitTime)

	err = dataset.Run(api, datasetFolder, datasetID, userID, accessionIDs)
	if err != nil {
		return err
	}

	slog.Info("dataset submission completed!")
	return nil
}
