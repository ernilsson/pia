package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use:   "pia",
	Short: "pia is a simple tool used to call and test web API:s",
	Run:   func(cmd *cobra.Command, args []string) {},
}

func Execute() {
	if err := root.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
