package helpers

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
)

type Command int

const (
	Unknown Command = iota
	Ingest
	Accession
	Dataset
)

func (c Command) String() string {
	switch c {
	case Ingest:
		return "ingest"
	case Accession:
		return "accession"
	case Dataset:
		return "dataset"
	default:
		return "unknown"
	}
}

var commandMap = map[string]Command{
	"ingest":    Ingest,
	"accession": Accession,
	"dataset":   Dataset,
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
	s.Suffix = " " + description
	s.Start()

	err := fn()
	s.Stop()

	if err != nil {
		fmt.Printf("❌ %s FAILED: %v\n", description, err)
		os.Exit(1)
	}
	fmt.Printf("✅ %s COMPLETE\n", description)
}

func ConfirmInputs(userID string, datasetFolder string, command Command, dryRun bool) {
	fmt.Println("\n====== Summary ======")
	fmt.Printf("User ID        : %s\n", userID)
	fmt.Printf("Dataset Folder : %s\n", datasetFolder)
	fmt.Printf("Command        : %s\n", command)
	fmt.Printf("Dry Run        : %t\n", dryRun)
	fmt.Println("=======================")

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Proceed? (Y/N): ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))

		if answer == "y" || answer == "yes" {
			break
		} else if answer == "n" || answer == "no" {
			fmt.Println("Aborted by user.")
			os.Exit(0)
		} else {
			fmt.Println("Please enter Y or N.")
		}
	}

}
