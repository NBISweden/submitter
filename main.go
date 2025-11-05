package main

import (
	"os"

	"github.com/NBISweden/submitter/cmd"
	_ "github.com/NBISweden/submitter/internal/accession"
	_ "github.com/NBISweden/submitter/internal/dataset"
	_ "github.com/NBISweden/submitter/internal/ingest"
	_ "github.com/NBISweden/submitter/internal/job"
	_ "github.com/NBISweden/submitter/internal/mail"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
