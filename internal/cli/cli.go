package cli

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/NBISweden/submitter/helpers"
)

type Inputs struct {
	Command     helpers.Command
	DryRun      bool
	ConfigFile  string
}

var ErrRequieredArguments = errors.New("Missing requiered input(s)")

func ParseArgs() (*Inputs, error) {
	inputs := &Inputs{}
	flag.BoolVar(&inputs.DryRun, "dry-run", true, "Used for running without executing impacting API calls, default=true")
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

