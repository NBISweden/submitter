package validate

import (
	"fmt"
	"os"

	"github.com/NBISweden/submitter/pkg/sdaclient"
)

type GitHubResponse struct {
	Name        string `json:"name"`
	DownloadURL string `json:"download_url"`
}

func ValidateSubmission(client *sdaclient.Client, dryRun bool) error {
	fmt.Println("Running validation")
	getXsdFiles()
	return nil
}

func getXsdFiles() error {
	fmt.Println("Creating directory for xsd files")
	err := os.MkdirAll("data/validator", os.ModePerm)
	if err != nil {
		return err
	}
	fmt.Println("Done!")
	return nil
}
