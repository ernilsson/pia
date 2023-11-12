package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use:   "pia",
	Short: "pia is a fast and easy to use API communicator for developers",
	Run:   func(cmd *cobra.Command, args []string) {},
}

func Execute() {
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
