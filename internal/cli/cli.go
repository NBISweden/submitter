package cli

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/NBISweden/submitter/helpers"
)

type Inputs struct {
	Command     helpers.Command
	DryRun      bool
	S3CmdConfig string
	ConfigFile  string
}

var ErrRequieredArguments = errors.New("Missing requiered input(s)")

func ParseArgs() (*Inputs, error) {
	inputs := &Inputs{}
	flag.BoolVar(&inputs.DryRun, "dry-run", true, "Used for running without executing impacting API calls, default=true")
	flag.StringVar(&inputs.S3CmdConfig, "s3config", "s3cmd.conf", "The s3cmd config file, default=s3cmd.config")
	flag.StringVar(&inputs.ConfigFile, "config", "config.yaml", "Path to configuration file, default=config.yaml")
	flag.Parse()

	args := flag.Args()
	cmd := args[0]
	inputs.Command = helpers.ParseCommand(cmd)
	if inputs.Command == helpers.Unknown {
		return inputs, fmt.Errorf("Command '%s' not valid, expecing one of [%s]\n", cmd, strings.Join(helpers.ValidCommands(), ", "))
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
