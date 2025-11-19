package main

import (
	"log/slog"
	"os"

	"github.com/NBISweden/submitter/cmd"
	_ "github.com/NBISweden/submitter/internal/accession"
	_ "github.com/NBISweden/submitter/internal/dataset"
	_ "github.com/NBISweden/submitter/internal/ingest"
	_ "github.com/NBISweden/submitter/internal/job"
	_ "github.com/NBISweden/submitter/internal/mail"
)

var version = "v1.0.1"

func main() {
	slog.Info("running", "version", version)
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
