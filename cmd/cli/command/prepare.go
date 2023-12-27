package command

import (
	"context"
	"github.com/ernilsson/pia/app"
	"github.com/spf13/cobra"
	"os"
)

var Prepare = &cobra.Command{
	Use:        "prepare",
	Aliases:    []string{"Prepare"},
	Short:      "prepares a request without executing it and writes the result to stdout",
	Args:       cobra.ExactArgs(1),
	ArgAliases: []string{"exchange configuration file"},
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := app.New()
		if err != nil {
			return err
		}
		ctx := context.Background()
		ex, err := a.LoadExchange(ctx, args[0], cmd)
		if err != nil {
			return err
		}
		req, err := a.Prepare(ctx, ex, cmd)
		if err != nil {
			return err
		}
		return WriteRequest(os.Stdout, req)
	},
}

func init() {
	Root.AddCommand(Prepare)
}
