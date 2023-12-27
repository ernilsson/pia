package app

import (
	"context"
	"github.com/ernilsson/pia/exchange"
	"github.com/ernilsson/pia/plugin"
	"github.com/ernilsson/pia/profile"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

type App struct {
	hooks      *plugin.HookService
	augmenters *plugin.AugmenterService
	profiles   profile.Store
}

func New() (*App, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	profiles := profile.Must(profile.NewFileStore(wd))
	hooks, err := plugin.NewHookService()
	if err != nil {
		return nil, err
	}
	augmenters, err := plugin.NewAugmenterService()
	if err != nil {
		return nil, err
	}
	return &App{
		hooks:      hooks,
		augmenters: augmenters,
		profiles:   profiles,
	}, nil
}

func (a *App) LoadExchange(ctx context.Context, base string, cmd *cobra.Command) (exchange.Exchange, error) {
	prof, err := a.profiles.LoadActive()
	if err != nil {
		return exchange.Exchange{}, err
	}
	processors, err := a.augmenters.ExchangePreProcessors(ctx, cmd)
	if err != nil {
		return exchange.Exchange{}, err
	}
	provider, err := exchange.DiscoveringFileProvider(base)
	if err != nil {
		return exchange.Exchange{}, err
	}
	return exchange.GetExchange(
		provider,
		append(processors, exchange.TemplatedConfiguration(prof))...,
	)
}

func (a *App) Prepare(ctx context.Context, ex exchange.Exchange, cmd *cobra.Command) (*http.Request, error) {
	processors, err := a.augmenters.BodyPreProcessors(ctx, cmd)
	if err != nil {
		return nil, err
	}
	if err := a.hooks.BeforeRequestPrepared(ctx, cmd, &ex); err != nil {
		return nil, err
	}
	prof, err := a.profiles.LoadActive()
	if err != nil {
		return nil, err
	}
	return exchange.NewRequest(
		ex,
		exchange.PreProcessedBody(
			append(processors, exchange.SubstitutionPreProcessor(prof))...,
		),
	)
}

func (a *App) Do(ctx context.Context, ex exchange.Exchange, cmd *cobra.Command) (*http.Response, error) {
	req, err := a.Prepare(ctx, ex, cmd)
	if err != nil {
		return nil, err
	}
	if err := a.hooks.BeforeRequestDispatched(ctx, cmd, &ex, req); err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if err := a.hooks.OnResponse(ctx, cmd, &ex, res); err != nil {
		return nil, err
	}
	return res, nil
}
