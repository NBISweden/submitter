package helpers

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/NBISweden/submitter/internal/accession"
	"github.com/NBISweden/submitter/internal/client"
)

func WaitForAccession(sdaclient *client.Client, target int, interval time.Duration, timeout time.Duration) ([]string, error) {
	deadline := time.Now().Add(timeout)
	for {
		paths, err := accession.GetVerifiedFilePaths(sdaclient)
		if err != nil {
			return nil, err
		}

		if len(paths) >= target {
			return paths, nil
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout reached, only got %d/%d files", len(paths), target)
		}
		slog.Info(fmt.Sprintf("[accession] found %d/%d files - waiting: internal: %s timeout: %s", len(paths), target, interval, timeout))
		time.Sleep(interval)
	}
}
