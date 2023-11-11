package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var verbose bool

var root = &cobra.Command{
	Use:   "pia",
	Short: "pia is a fast and easy to use API communicator for developers",
	Run:   func(cmd *cobra.Command, args []string) {},
}

func init() {
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enables verbose logging")
}

func Execute() {
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
