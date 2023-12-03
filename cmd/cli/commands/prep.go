package commands

import (
	"fmt"
	"github.com/ernilsson/pia/exchange"
	"github.com/ernilsson/pia/plug"
	"github.com/ernilsson/pia/profile"
	"github.com/spf13/cobra"
	"io"
	"os"
	"strings"
)

var prep = &cobra.Command{
	Use:     "prepare",
	Aliases: []string{"prep"},
	Short:   "prepares a request without executing it and writes the result to stdout",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prof, err := profile.UnmarshalActive()
		if err != nil {
			return err
		}
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		vars, err := parseVariables(cmd)
		if err != nil {
			return err
		}
		target := args[0]
		ex, err := exchange.GetExchange(exchange.FileProvider(fmt.Sprintf("%s/%s", wd, target)), prof, vars)
		if err != nil {
			return err
		}
		req, err := exchange.NewRequest(ex.Request, exchange.TemplatedBody(prof, exchange.VariableSet(vars)))
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

		fmt.Printf("URL: %s\n", req.URL)
		fmt.Printf("Method: %s\n", req.Method)
		for key, v := range req.Header {
			fmt.Printf("%s: %s\n", key, v[0])
		}
		fmt.Println()
		parsed, err := io.ReadAll(req.Body)
		if err != nil {
			return err
		}
		fmt.Println(string(parsed))
		return nil
	},
}

func parseVariables(cmd *cobra.Command) (map[string]any, error) {
	raw, err := cmd.Flags().GetStringSlice("var")
	if err != nil {
		return nil, err
	}
	vars := make(map[string]any)
	for _, kv := range raw {
		split := strings.Split(kv, "=")
		vars[split[0]] = split[1]
	}
	return vars, nil
}

func init() {
	prep.Flags().StringSlice("var", nil, "sets a variable for the request body")
	root.AddCommand(prep)
}
