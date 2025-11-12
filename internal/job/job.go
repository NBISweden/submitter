package job

import (
	"log/slog"
	"time"

	"github.com/NBISweden/submitter/cmd"
	"github.com/NBISweden/submitter/helpers"
	"github.com/NBISweden/submitter/internal/accession"
	"github.com/NBISweden/submitter/internal/client"
	"github.com/NBISweden/submitter/internal/config"
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
		conf, err := config.NewConfig(configPath)
		if err != nil {
			return err
		}
		err = runJob(conf)
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

func runJob(conf config.Config) error {
	api := client.New(conf)
	filesCount, err := ingest.IngestFiles(api, false)
	if err != nil {
		return err
	}
	//TODO: Look at this logic. Goal: remove the part where we store data on disk, keep it in memory for the job
	_, err = helpers.WaitForAccession(api, filesCount, 5*time.Minute, 24*time.Hour)
	if err != nil {
		return err
	}
	err = accession.CreateAccessionIDs(api, conf)
	if err != nil {
		return err
	}

	// We give some time for the SDA backend to process our accession ids. During test-runs it's been fine with 2 minutes
	waitTime := 2 * time.Minute
	slog.Info("waiting before sending dataset creation request", "delay", waitTime)
	time.Sleep(waitTime)

	err = dataset.CreateDataset(api, conf)
	if err != nil {
		return err
	}

	slog.Info("dataset submission %s completed!", "datasetID", api.DatasetID)
	return nil
}
