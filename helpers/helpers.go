package helpers

import (
	"bufio"
	"strings"
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
)

var AllowedCommands = []string{"ingest", "accession", "dataset"}


func IsCommandAllowed(cmd string) error{
	// Construct a set of the list of allowed commands for lookup
	allowedSet := make(map[string]struct{}, len(AllowedCommands))
	for _, v := range AllowedCommands {
		allowedSet[v] = struct{}{}
	}
	if _, ok := allowedSet[cmd]; !ok {
		return fmt.Errorf("Command '%s' not allowed, expecting one of [%s]", cmd, strings.Join(AllowedCommands, ", "))
	}
	return nil

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

func ConfirmInputs(userID string, datasetFolder string, command string, dryRun bool) {
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
