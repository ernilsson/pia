package commands

import (
	"github.com/ernilsson/pia/exchange"
	"github.com/ernilsson/pia/profile"
	"github.com/spf13/cobra"
	"os"
	"path"
)

var prep = &cobra.Command{
	Use:        "prepare",
	Aliases:    []string{"prep"},
	Short:      "prepares a request without executing it and writes the result to stdout",
	Args:       cobra.ExactArgs(1),
	ArgAliases: []string{"exchange configuration file"},
	RunE: func(cmd *cobra.Command, args []string) error {
		raw, err := cmd.Flags().GetStringSlice("var")
		if err != nil {
			return err
		}
		vars := MustParse(ParseKeyValues(raw))

		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		filepath, err := DiscoverExchangeFile(path.Join(wd, args[0]))
		if err != nil {
			return err
		}
		store := profile.NewFileStore(wd)
		prof, err := store.LoadActive()
		if err != nil {
			return err
		}
		ex, err := exchange.GetExchange(
			exchange.FileProvider(filepath),
			exchange.TemplatedConfiguration(prof),
		)
		if err != nil {
			return err
		}
		ex.ConfigRoot = path.Dir(filepath)
		req, err := exchange.NewRequest(ex, exchange.TemplatedBody(prof, exchange.VariableSet(vars)))
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
