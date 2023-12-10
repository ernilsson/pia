package commands

import (
	"github.com/ernilsson/pia/exchange"
	"github.com/ernilsson/pia/profile"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"path"
)

var do = &cobra.Command{
	Use:        "do",
	Short:      "executes a request and writes the response to stdout",
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
		filepath, err := DiscoverExchangeFile(path.Join("wd", args[0]))
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

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		return WriteResponse(os.Stdout, res)
	},
}

func init() {
	do.Flags().StringSlice("var", nil, "set variables for the request body, ex: --var id=1")
	root.AddCommand(do)
}
