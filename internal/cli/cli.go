package cli

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/NBISweden/submitter/helpers"
	// "github.com/NBISweden/submitter/helpers"
)

type Inputs struct {
	Command       string
	DryRun        bool
	APIHost       string
	S3CmdConfig   string
	UserID        string
	DatasetID     string
	DatasetFolder string
}

var ErrRequieredArguments = errors.New("Missing requiered input(s)")

func ParseArgs() (*Inputs, error) {
	inputs := &Inputs{}
	flag.BoolVar(&inputs.DryRun, "dry-run", true, "Used for running without executing impacting API calls, default=true")
	flag.StringVar(&inputs.APIHost, "api-host", "https://api.bp.nbis.se", "The Big Picture API URL, default=https://api.bp.nbis.se")
	flag.StringVar(&inputs.S3CmdConfig, "config", "s3cmd.conf", "The s3cmd config file, default=s3cmd.config")
	flag.StringVar(&inputs.UserID, "user-id", "", "The User ID of the uploader / submitter (requiered)")
	flag.StringVar(&inputs.DatasetID, "dataset-id", "", "The ID of the uploaded dataset (requiered)")
	flag.StringVar(&inputs.DatasetFolder, "dataset-folder", "", "The folder in s3inbox where the uploaded files reside (requiered)")

	flag.Parse()

	args := flag.Args()
	cmd := args[0]
	err := helpers.IsCommandAllowed(cmd)
	if err != nil {
		return inputs, err
	}

	inputs.Command = cmd

	var missingArgs []string

	if inputs.UserID == "" {
		missingArgs = append(missingArgs, "user-id")
	}

	if inputs.DatasetID == "" {
		missingArgs = append(missingArgs, "dataset-id")
	}

	if inputs.DatasetFolder == "" {
		missingArgs = append(missingArgs, "dataset-folder")
	}

	if len(missingArgs) > 0 {
		return inputs, fmt.Errorf("%w: %s", ErrRequieredArguments, strings.Join(missingArgs, ", "))
	}

	return inputs, nil
}

func (i *Inputs) GetAccessToken() (string, error) {
	file, err := os.Open(i.S3CmdConfig)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "access_token" {
			return value, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", nil
	}
	return "", fmt.Errorf("access_token not found in %s", i.S3CmdConfig)
}

func (i *Inputs) Validation() {
	i.UserID = strings.ReplaceAll(i.UserID, "@", "_")
}
