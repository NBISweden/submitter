package main

import (
	"fmt"
	"os"

	"github.com/NBISweden/submitter/cmd/ingest"
	"github.com/NBISweden/submitter/internal/cli"
)

func main() {
	inputs, err := cli.ParseArgs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	inputs.Validation()
	token, err := inputs.GetAccessToken()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = ingest.IngestFiles(token, inputs.APIHost, inputs.UserID, inputs.DatasetFolder, inputs.DryRun)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
