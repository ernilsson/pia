package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

var do = &cobra.Command{
	Use:  "do",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vars, err := cmd.Flags().GetStringSlice("var")
		if err != nil {
			return err
		}
		vs, err := ParseKeyValues(vars)
		if err != nil {
			return err
		}
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		req, err := PrepareRequest(fmt.Sprintf("%s/%s", wd, args[0]), vs)
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
	do.Flags().StringSlice("var", nil, "sets a variable for the request body")
	root.AddCommand(do)
}
