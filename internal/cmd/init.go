package cmd

import (
	"os"

	"github.com/ernilsson/pia/environment"
	"github.com/spf13/cobra"
)

var initialize = &cobra.Command{
	Use:   "init",
	Short: "Initializes a new pia project in your current working directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		return environment.Bootstrap(wd)
	},
}

func init() {
	root.AddCommand(initialize)
}
