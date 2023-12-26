package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = &cobra.Command{
	Use:   "version",
	Short: "currently installed pia Version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pia pre-release")
	},
}

func init() {
	Root.AddCommand(Version)
}
