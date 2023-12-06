package commands

import (
	"fmt"
	"github.com/ernilsson/pia/exchange"
	"github.com/ernilsson/pia/plug"
	"github.com/ernilsson/pia/profile"
	"github.com/spf13/cobra"
	"os"
)

var prep = &cobra.Command{
	Use:     "prepare",
	Aliases: []string{"prep"},
	Short:   "prepares a request without executing it and writes the result to stdout",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		store := profile.NewFileStore(wd)
		prof, err := store.LoadActive()
		if err != nil {
			return err
		}
		ex, err := exchange.GetExchange(
			exchange.FileProvider(fmt.Sprintf("%s/%s", wd, args[0])),
			exchange.TemplatedConfiguration(prof),
		)
		if err != nil {
			return err
		}
		vars, err := cmd.Flags().GetStringSlice("var")
		if err != nil {
			return err
		}
		vs, err := ParseKeyValues(vars)
		if err != nil {
			return err
		}
		req, err := exchange.NewRequest(ex.Request, exchange.TemplatedBody(prof, exchange.VariableSet(vs)))
		if err != nil {
			return err
		}
		hooks, err := plug.LoadRequestHooks("./")
		if err != nil {
			return err
		}
		for _, hook := range hooks {
			if err := hook.OnRequest(ex, req); err != nil {
				return err
			}
		}

		return WriteRequest(os.Stdout, req)
	},
}

func init() {
	prep.Flags().StringSlice("var", nil, "sets a variable for the request body")
	root.AddCommand(prep)
}
