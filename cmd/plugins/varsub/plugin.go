package main

import (
	"context"
	"github.com/ernilsson/pia/cmd/cli/command"
	"github.com/ernilsson/pia/cmd/plugins/varsub/internal"
	"github.com/ernilsson/pia/exchange"
	"github.com/spf13/cobra"
)

func ExchangePreProcessorFactory(ctx context.Context, cmd *cobra.Command) (exchange.PreProcessor, error) {
	pairs, err := cmd.Flags().GetStringSlice("variable")
	if err != nil {
		return nil, err
	}
	vars, err := internal.Split(pairs)
	if err != nil {
		return nil, err
	}
	return func(raw []byte) ([]byte, error) {
		return vars.SubstituteLines(raw)
	}, nil
}

func BodyPreProcessorFactory(ctx context.Context, cmd *cobra.Command) (exchange.PreProcessor, error) {
	pairs, err := cmd.Flags().GetStringSlice("variable")
	if err != nil {
		return nil, err
	}
	vars, err := internal.Split(pairs)
	if err != nil {
		return nil, err
	}
	return func(raw []byte) ([]byte, error) {
		return vars.SubstituteLines(raw)
	}, nil
}

func OnInit() error {
	command.Prepare.Flags().StringSliceP("variable", "v", nil, "sets a variable")
	command.Do.Flags().StringSliceP("variable", "v", nil, "sets a variable")
	return nil
}
