package hook

import (
	"context"
	"errors"
	"github.com/ernilsson/pia/exchange"
	"github.com/spf13/cobra"
	"net/http"
)

const dir = "/etc/pia/plugins"

func OnInit() error {
	plugins, err := load(dir)
	if err != nil {
		return err
	}
	for _, p := range plugins {
		h, err := p.Lookup("OnInit")
		if err != nil {
			continue
		}
		hook, ok := h.(func() error)
		if !ok {
			return errors.New("invalid OnInit hook installed")
		}
		if err := hook(); err != nil {
			return err
		}
	}
	return nil
}

func GetExchangePreProcessors(ctx context.Context, cmd *cobra.Command) ([]exchange.PreProcessor, error) {
	plugins, err := load(dir)
	if err != nil {
		return nil, err
	}
	processors := make([]exchange.PreProcessor, 0)
	for _, p := range plugins {
		h, err := p.Lookup("ExchangePreProcessorFactory")
		if err != nil {
			continue
		}
		factory, ok := h.(func(context.Context, *cobra.Command) (exchange.PreProcessor, error))
		if !ok {
			return nil, errors.New("invalid ExchangePreProcessorFactory installed")
		}
		processor, err := factory(ctx, cmd)
		if err != nil {
			return nil, err
		}
		processors = append(processors, processor)
	}
	return processors, nil
}

func GetBodyPreProcessors(ctx context.Context, cmd *cobra.Command) ([]exchange.PreProcessor, error) {
	plugins, err := load(dir)
	if err != nil {
		return nil, err
	}
	processors := make([]exchange.PreProcessor, 0)
	for _, p := range plugins {
		h, err := p.Lookup("BodyPreProcessorFactory")
		if err != nil {
			continue
		}
		factory, ok := h.(func(context.Context, *cobra.Command) (exchange.PreProcessor, error))
		if !ok {
			return nil, errors.New("invalid BodyPreProcessorFactory installed")
		}
		processor, err := factory(ctx, cmd)
		if err != nil {
			return nil, err
		}
		processors = append(processors, processor)
	}
	return processors, nil
}

func BeforeRequestPrepared(ctx context.Context, cmd *cobra.Command, ex *exchange.Exchange) error {
	plugins, err := load(dir)
	if err != nil {
		return err
	}
	for _, p := range plugins {
		h, err := p.Lookup("BeforeRequestPrepared")
		if err != nil {
			continue
		}
		hook, ok := h.(func(context.Context, *cobra.Command, *exchange.Exchange) error)
		if !ok {
			return errors.New("invalid BeforeRequestPrepared hook installed")
		}
		if err := hook(ctx, cmd, ex); err != nil {
			return err
		}
	}
	return nil
}

func BeforeRequestDispatched(ctx context.Context, cmd *cobra.Command, ex *exchange.Exchange, r *http.Request) error {
	plugins, err := load(dir)
	if err != nil {
		return err
	}
	for _, p := range plugins {
		h, err := p.Lookup("BeforeRequestDispatched")
		if err != nil {
			continue
		}
		hook, ok := h.(func(context.Context, *cobra.Command, *exchange.Exchange, *http.Request) error)
		if !ok {
			return errors.New("invalid BeforeRequestDispatched hook installed")
		}
		if err := hook(ctx, cmd, ex, r); err != nil {
			return err
		}
	}
	return nil
}

func OnResponse(ctx context.Context, cmd *cobra.Command, ex *exchange.Exchange, r *http.Response) error {
	plugins, err := load(dir)
	if err != nil {
		return err
	}
	for _, p := range plugins {
		h, err := p.Lookup("OnResponse")
		if err != nil {
			continue
		}
		hook, ok := h.(func(context.Context, *cobra.Command, *exchange.Exchange, *http.Response) error)
		if !ok {
			return errors.New("invalid OnResponse hook installed")
		}
		if err := hook(ctx, cmd, ex, r); err != nil {
			return err
		}
	}
	return nil
}
