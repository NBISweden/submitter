package cli

import (
	"flag"
	"fmt"
	"strings"

	"github.com/NBISweden/submitter/helpers"
)

type Inputs struct {
	Command    helpers.Command
	DryRun     bool
	ConfigFile string
}

func ParseArgs() (*Inputs, error) {
	inputs := &Inputs{}
	flag.BoolVar(&inputs.DryRun, "dry-run", true, "Used for running without executing impacting API calls, default=true")
	flag.StringVar(&inputs.ConfigFile, "config", "config.yaml", "Path to configuration file, default=config.yaml")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		return inputs, fmt.Errorf("argument not supplied, need one of [%s]", strings.Join(helpers.ValidCommands(), ", "))
	}

	cmd := args[0]
	inputs.Command = helpers.ParseCommand(cmd)
	if inputs.Command == helpers.Unknown {
		return inputs, fmt.Errorf("argument '%s' not valid, expecing one of [%s]", cmd, strings.Join(helpers.ValidCommands(), ", "))
	}

	return inputs, nil
}
