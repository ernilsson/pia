package commands

import (
	"context"
	"github.com/ernilsson/pia/exchange"
	"github.com/ernilsson/pia/hook"
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
		ctx := context.Background()

		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		filepath, err := DiscoverExchangeFile(path.Join(wd, args[0]))
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
		if err := hook.BeforeRequestPrepared(context.Background(), cmd, &ex); err != nil {
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
		if err := hook.BeforeRequestDispatched(context.Background(), cmd, &ex, req); err != nil {
			return err
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		if err := hook.OnResponse(context.Background(), cmd, &ex, res); err != nil {
			return err
		}
		return WriteResponse(os.Stdout, res)
	},
}

func init() {
	Root.AddCommand(do)
}
