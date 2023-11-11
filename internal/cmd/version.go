package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = &cobra.Command{
	Use:   "version",
	Short: "currently installed pia version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pia pre-release 0.0.1")
	},
}

func init() {
	root.AddCommand(version)
}
