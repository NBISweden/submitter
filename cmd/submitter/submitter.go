package main

import (
	"fmt"
	"log"
	"time"

	"github.com/NBISweden/submitter/helpers"
	"github.com/NBISweden/submitter/internal/accession"
	"github.com/NBISweden/submitter/internal/cli"
	"github.com/NBISweden/submitter/internal/config"
	"github.com/NBISweden/submitter/internal/dataset"
	"github.com/NBISweden/submitter/internal/ingest"
	"github.com/NBISweden/submitter/internal/mail"
	"github.com/NBISweden/submitter/pkg/sdaclient"
)

func main() {
	start := time.Now()
	inputs, err := cli.ParseArgs()
	if err != nil {
		log.Fatalf("failed to parse args: %v", err)
	}

	conf, err := config.NewConfig(inputs.ConfigFile)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	token, err := conf.GetAccessToken()
	if err != nil {
		log.Fatalf("failed to retrieve access token: %v", err)
	}

	client := sdaclient.NewClient(token, conf.APIHost, conf.UserID, conf.DatasetFolder, conf.DatasetID)

	if err := runCommand(inputs.Command, client, conf, inputs.DryRun); err != nil {
		log.Fatalf("command %q failed: %v", inputs.Command, err)
	}
	elapsed := time.Since(start)
	fmt.Printf("[Submitter] Execution time: %s\n", elapsed)
}

func runCommand(cmd helpers.Command, client *sdaclient.Client, conf *config.Config, dryRun bool) error {
	switch cmd {
	case helpers.Ingest:
		return ingestFiles(client, dryRun)
	case helpers.Accession:
		return createAccession(client, dryRun)
	case helpers.Dataset:
		return createDataset(client, dryRun)
	case helpers.Mail:
		return sendMail(conf)
	case helpers.All:
		return runAll(client, conf, dryRun)
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

func ingestFiles(client *sdaclient.Client, dryRun bool) error {
	_, err := ingest.IngestFiles(client, dryRun)
	return err
}

func createAccession(client *sdaclient.Client, dryRun bool) error {
	return accession.CreateAccessionIDs(client, dryRun)
}

func createDataset(client *sdaclient.Client, dryRun bool) error {
	return dataset.CreateDataset(client, dryRun)
}

func sendMail(conf *config.Config) error {
	m := mail.Configure(conf)

	for _, recipient := range []string{"BigPicture", "Jarno", "Submitter"} {
		if err := m.Notify(recipient); err != nil {
			return fmt.Errorf("failed to notify %s: %w", recipient, err)
		}
	}
	return nil
}

func runAll(client *sdaclient.Client, conf *config.Config, dryRun bool) error {
	if dryRun {
		fmt.Println("dry run true enabled, exiting...")
		return nil
	}
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

	// We give some time for the SDA backend to process our accession ids. During test-runs it's been fine with 2 minutes. Might need more.
	delay := 2 * time.Minute
	fmt.Printf("[Submitter] Waiting %s before sending dataset creation request\n", delay)
	time.Sleep(delay)

	err = dataset.CreateDataset(client, false)
	if err != nil {
		return err
	}

	fmt.Printf("[Submitter] Dataset Submission %s completed!\n", conf.DatasetID)
	return nil
}
