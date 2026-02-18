package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "bpctl",
	Short:        "Runs dataset submissions",
	Long:         `Big Picture Control (bpctl) is a tool that solves domain specific problems for the Big Picture project, it can be used to do data ingestion, create accession ids and dataset mapping through the Sensetive Data Archive (SDA) and it's corresponding API. bpctl can be used either as a command line tool for a user / admin in Big Picture to do specific parts of administration process, or it can be run as a job in kubernetes executing a end-to-end flow of the same administrative tasks.`,
	SilenceUsage: true,
}

func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func AddCommand(command *cobra.Command) {
	rootCmd.AddCommand(command)
}

func AddVersion(v string) {
	rootCmd.Version = v
}
