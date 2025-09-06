package main

import (
	"fmt"
	"net/http"

	"github.com/NBISweden/submitter/helpers"
	"github.com/NBISweden/submitter/internal/accession"
	"github.com/NBISweden/submitter/internal/cli"
	"github.com/NBISweden/submitter/internal/config"
	"github.com/NBISweden/submitter/internal/ingest"
	"github.com/NBISweden/submitter/internal/mail"
	"github.com/NBISweden/submitter/pkg/sdaclient"
)

func main() {
	var inputs *cli.Inputs
	var token string
	var sdaClient *sdaclient.Client
	var conf *config.Config
	var err error

	helpers.RunStep("Parsing arguments", func() error {
		inputs, err = cli.ParseArgs()
		if err != nil {
			return err
		}
		return nil
	})

	helpers.RunStep("Reading Config", func() error {
		conf, err = config.NewConfig(inputs.ConfigFile)
		if err != nil {
			return err
		}
		return nil
	})

	helpers.RunStep("Getting Access Token", func() error {
		token, err = inputs.GetAccessToken()
		return err
	})

	helpers.RunStep("Creating SDA Client", func() error {
		sdaClient = &sdaclient.Client{
			AccessToken:   token,
			APIHost:       conf.APIHost,
			UserID:        conf.UserID,
			DatasetFolder: conf.DatasetFolder,
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
			fmt.Println(conf)
			return nil
		})
	}

	if inputs.Command == helpers.Mail {
		helpers.RunStep("Sending email notification", func() error {
			m := mail.Configure(conf)

			err := m.Notify("BigPicture")
			if err != nil {
				return err
			}

			err = m.Notify("Jarno")
			if err != nil {
				return err
			}

			err = m.Notify("Submitter")
			if err != nil {
				return err
			}

			return err
		})
	}

}
