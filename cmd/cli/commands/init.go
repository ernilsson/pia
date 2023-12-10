package commands

import (
	"github.com/ernilsson/pia/profile"
	"github.com/spf13/cobra"
	"os"
)

var initialize = &cobra.Command{
	Use:        "initialize",
	Aliases:    []string{"init"},
	Short:      "initializes a new pia project in the current working directory",
	Args:       cobra.ExactArgs(1),
	ArgAliases: []string{"initial profile name"},
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		store := profile.Must(profile.NewFileStore(wd))
		prof := profile.New(args[0])
		if err := store.Save(prof); err != nil {
			return err
		}
		return store.SetActive(args[0])
	},
}

func init() {
	root.AddCommand(initialize)
}
