package command

import (
	"context"
	"github.com/ernilsson/pia/exchange"
	"github.com/ernilsson/pia/hook"
	"github.com/ernilsson/pia/profile"
	"github.com/spf13/cobra"
	"os"
	"path"
)

var Prepare = &cobra.Command{
	Use:        "prepare",
	Aliases:    []string{"Prepare"},
	Short:      "prepares a request without executing it and writes the result to stdout",
	Args:       cobra.ExactArgs(1),
	ArgAliases: []string{"exchange configuration file"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		filepath, err := DiscoverExchangeFile(path.Join(wd, args[0]))
		if err != nil {
			return err
		}
		store := profile.Must(profile.NewFileStore(wd))
		prof, err := store.LoadActive()
		if err != nil {
			return err
		}
		processors, err := hook.GetExchangePreProcessors(ctx, cmd)
		if err != nil {
			return err
		}
		ex, err := exchange.GetExchange(
			exchange.FileProvider(filepath),
			append(processors, exchange.TemplatedConfiguration(prof))...,
		)
		if err != nil {
			return err
		}
		ex.ConfigRoot = path.Dir(filepath)
		if err := hook.BeforeRequestPrepared(ctx, cmd, &ex); err != nil {
			return err
		}
		processors, err = hook.GetBodyPreProcessors(ctx, cmd)
		if err != nil {
			return err
		}
		req, err := exchange.NewRequest(
			ex,
			exchange.PreProcessedBody(
				append(processors, exchange.SubstitutionPreProcessor(prof))...,
			),
		)
		if err != nil {
			return err
		}
		if err := hook.BeforeRequestDispatched(ctx, cmd, &ex, req); err != nil {
			return err
		}
		return WriteRequest(os.Stdout, req)
	},
}

func init() {
	Root.AddCommand(Prepare)
}
