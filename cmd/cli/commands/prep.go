package commands

import (
	"github.com/spf13/cobra"
	"os"
)

var prep = &cobra.Command{
	Use:        "prepare",
	Aliases:    []string{"prep"},
	Short:      "prepares a request without executing it and writes the result to stdout",
	Args:       cobra.ExactArgs(1),
	ArgAliases: []string{"exchange configuration file"},
	RunE: func(cmd *cobra.Command, args []string) error {
		vars, err := cmd.Flags().GetStringSlice("var")
		if err != nil {
			return err
		}
		vs, err := ParseKeyValues(vars)
		if err != nil {
			return err
		}
		filepath, err := DiscoverExchangeFile(args[0])
		if err != nil {
			return err
		}
		req, err := PrepareRequest(filepath, vs)
		if err != nil {
			return err
		}
		return WriteRequest(os.Stdout, req)
	},
}

func init() {
	prep.Flags().StringSlice("var", nil, "set variables for the request body, ex: --var id=1")
	root.AddCommand(prep)
}
