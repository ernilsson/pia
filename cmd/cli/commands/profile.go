package commands

import (
	"github.com/ernilsson/pia/profile"
	"github.com/spf13/cobra"
	"os"
)

var prof = &cobra.Command{
	Use:        "profile",
	Aliases:    []string{"prof"},
	SuggestFor: nil,
	Short:      "manages the currently selected profile",
}

var set = &cobra.Command{
	Use:        "set",
	SuggestFor: nil,
	Short:      "sets the active profile for the current project",
	Args:       cobra.ExactArgs(1),
	ArgAliases: []string{"name"},
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		name := args[0]
		if err := profile.SetActiveProfileName(wd, name); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	prof.AddCommand(set)
	root.AddCommand(prof)
}
