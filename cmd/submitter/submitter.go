package main

import (
	"fmt"
	"net/http"

	"github.com/NBISweden/submitter/helpers"
	"github.com/NBISweden/submitter/internal/accession"
	"github.com/NBISweden/submitter/internal/cli"
	"github.com/NBISweden/submitter/internal/ingest"
	"github.com/NBISweden/submitter/pkg/sdaclient"
)

func main() {
	var inputs *cli.Inputs
	var token string
	var sdaClient *sdaclient.Client

	helpers.RunStep("Parsing arguments", func() error {
		var err error
		inputs, err = cli.ParseArgs()
		if err != nil {
			return err
		}
		inputs.Validation()
		return nil
	})
	helpers.ConfirmInputs(inputs.UserID, inputs.DatasetFolder, inputs.Command, inputs.DryRun)

	helpers.RunStep("Getting Access Token", func() error {
		var err error
		token, err = inputs.GetAccessToken()
		return err
	})

	helpers.RunStep("Creating SDA Client", func() error {
		sdaClient = &sdaclient.Client{
			AccessToken:   token,
			APIHost:       inputs.APIHost,
			UserID:        inputs.UserID,
			DatasetFolder: inputs.DatasetFolder,
			HTTPClient:    http.DefaultClient,
		}
		return nil
	})

	if inputs.Command == helpers.Ingest {
		helpers.RunStep("Ingesting Files", func() error {
			return ingest.IngestFiles(sdaClient, inputs.DryRun)
		})
	}

	if inputs.Command == helpers.Accession {
		helpers.RunStep("Creating Accession IDs", func() error {
			return accession.CreateAccessionIDs(sdaClient, "fileIDs.txt", inputs.DryRun)
		})
	}

	if inputs.Command == helpers.Dataset {
		helpers.RunStep("Creating Dataset", func() error {
			fmt.Println("Dataset created")
			return nil
		})
	}

	if inputs.Command == helpers.Mail {
		helpers.RunStep("Sending email notification", func() error {
			return nil
		})
	}

}
