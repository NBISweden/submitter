package job

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/NBISweden/sda-bpctl/cmd"
	"github.com/NBISweden/sda-bpctl/internal/accession"
	"github.com/NBISweden/sda-bpctl/internal/client"
	"github.com/NBISweden/sda-bpctl/internal/config"
	"github.com/NBISweden/sda-bpctl/internal/dataset"
	"github.com/NBISweden/sda-bpctl/internal/ingest"
	"github.com/NBISweden/sda-bpctl/internal/landingpage"
	"github.com/NBISweden/sda-bpctl/internal/mail"
	"github.com/spf13/cobra"
)

var configPath string

var jobCmd = &cobra.Command{
	Use:   "job <expectedFiles>",
	Short: "Runs all dataset submission steps in order",
	Long:  `Runs all dataset submission steps in order (ingestion -> accession -> dataset) takes a integer value representing the expected number of files to be included in the finalized dataset as argument. When a dataset is completed it ends with sending mail notifications and moving landing pages`,

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
	jobCmd.Flags().StringVarP(&configPath, "config", "c", "config.yaml", "Path to configuration file")
}

func runJob() error {
	cfg, err := config.NewConfig(configPath)
	if err != nil {
		return err
	}

	pollRate := time.Minute * time.Duration(cfg.PollRate)
	timeout := time.Minute * time.Duration(cfg.Timeout)
	datasetFolder := cfg.DatasetFolder
	datasetID := cfg.DatasetID
	userID := cfg.UserID
	dataDirectory := cfg.JobDataDirectory

	slog.Info("dispatching job", "dataset_folder", datasetFolder, "dataset_id", datasetID, "userID", userID)

	api, err := client.New(cfg)
	if err != nil {
		return err
	}

	filesCount, err := ingest.Run(api, datasetFolder, userID)
	if err != nil {
		return err
	}

	_, err = api.WaitForAccession(filesCount, pollRate, timeout)
	if err != nil {
		return err
	}

	accession.DataDirectory = dataDirectory
	accessionIDs, err := accession.Run(api, datasetFolder, userID)
	if err != nil {
		return err
	}

	nrAccessionIDs := len(accessionIDs)
	if nrAccessionIDs != filesCount {
		return fmt.Errorf("accession did not return the expected number of files, got %d expected %d", nrAccessionIDs, filesCount)
	}

	waitTime := 10 * time.Minute
	slog.Info("waiting before sending dataset creation request", "delay", waitTime)
	time.Sleep(waitTime)

	dataset.DataDirectory = dataDirectory
	err = dataset.Run(api, datasetFolder, datasetID, userID, accessionIDs)
	if err != nil {
		return err
	}

	err = landingpage.Run(cfg)
	if err != nil {
		slog.Warn("could not complete landingpage", "err", err)
	}

	mail.DataDirectory = dataDirectory
	err = mail.Run(cfg)
	if err != nil {
		slog.Warn("could not complete mail notifications", "err", err)
	}

	slog.Info("dataset submission completed!")
	return nil
}
