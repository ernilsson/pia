package command

import (
	"context"
	"github.com/ernilsson/pia/app"
	"github.com/spf13/cobra"
	"os"
)

var Do = &cobra.Command{
	Use:        "do",
	Short:      "executes a request and writes the response to stdout",
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
		res, err := a.Do(ctx, ex, cmd)
		if err != nil {
			return err
		}
		return WriteResponse(os.Stdout, res)
	},
}

func init() {
	Root.AddCommand(Do)
}
