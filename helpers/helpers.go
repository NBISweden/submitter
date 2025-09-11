package helpers

import (
	"fmt"
	"time"

	"github.com/NBISweden/submitter/internal/accession"
	"github.com/NBISweden/submitter/pkg/sdaclient"
)

type Command int

const (
	Unknown Command = iota
	Ingest
	Accession
	Dataset
	Mail
	All
)

func (c Command) String() string {
	switch c {
	case Ingest:
		return "ingest"
	case Accession:
		return "accession"
	case Dataset:
		return "dataset"
	case Mail:
		return "mail"
	case All:
		return "all"
	default:
		return "unknown"
	}
}

var commandMap = map[string]Command{
	"ingest":    Ingest,
	"accession": Accession,
	"dataset":   Dataset,
	"mail":      Mail,
	"all":       All,
}

func ParseCommand(s string) Command {
	if cmd, ok := commandMap[s]; ok {
		return cmd
	}
	return Unknown
}

func ValidCommands() []string {
	cmds := make([]string, 0, len(commandMap))
	for k := range commandMap {
		cmds = append(cmds, k)
	}
	return cmds
}

func WaitForAccession(client *sdaclient.Client, target int, interval time.Duration, timeout time.Duration) ([]string, error) {
	deadline := time.Now().Add(timeout)
	for {
		paths, err := accession.GetVerifiedFilePaths(client)
		if err != nil {
			return nil, err
		}

		if len(paths) >= target {
			return paths, nil
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout reached, only got %d/%d files", len(paths), target)
		}
		fmt.Printf("[Accession] Found %d files, waiting for %d files. Interval: %s Timeout: %s\n", len(paths), target, interval, timeout)
		time.Sleep(interval)
	}
}
