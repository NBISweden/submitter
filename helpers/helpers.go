package helpers

import (
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
)

type Command int

const (
	Unknown Command = iota
	Ingest
	Accession
	Dataset
	Mail
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
		return "Mail"
	default:
		return "unknown"
	}
}

var commandMap = map[string]Command{
	"ingest":    Ingest,
	"accession": Accession,
	"dataset":   Dataset,
	"mail":      Mail,
	"unknown":   Unknown,
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

func RunStep(description string, fn func() error) {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Color("cyan")
	s.Suffix = " " + description + "\n"
	s.Start()

	err := fn()
	s.Stop()

	if err != nil {
		fmt.Printf("❌ %s FAILED: %v\n", description, err)
		os.Exit(1)
	}
	fmt.Printf("✅ %s COMPLETE\n", description)
}
