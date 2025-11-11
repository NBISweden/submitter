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

var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "Runs all dataset submission steps as a 'job'",
	Long:  "Runs all dataset submission steps as a 'job' (ingestion, accession, dataset)",
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := config.NewConfig()
		if err != nil {
			return err
		}
		sdaclient := client.NewClient(*conf)
		err = runJob(sdaclient)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	cmd.AddCommand(jobCmd)
}

func runJob(client *client.Client) error {
	filesCount, err := ingest.IngestFiles(client, false)
	if err != nil {
		return err
	}
	_, err = helpers.WaitForAccession(client, filesCount, 5*time.Minute, 24*time.Hour)
	if err != nil {
		return err
	}
	err = accession.CreateAccessionIDs(client, false)
	if err != nil {
		return err
	}

	// We give some time for the SDA backend to process our accession ids. During test-runs it's been fine with 2 minutes
	waitTime := 2 * time.Minute
	slog.Info("Waiting before sending dataset creation request", "delay", waitTime)
	time.Sleep(waitTime)

	err = dataset.CreateDataset(client, false)
	if err != nil {
		return err
	}

	slog.Info("Dataset Submission %s completed!", "DatasetID", client.DatasetID)
	return nil
}
