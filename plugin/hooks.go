package plugin

import (
	"context"
	"errors"
	"github.com/ernilsson/pia/exchange"
	"github.com/spf13/cobra"
	"net/http"
	"plugin"
)

const dir = "/etc/pia/plugins"

type IntermediateResult[T HookService | AugmenterService] struct {
	result T
	err    error
}

func (ir IntermediateResult[T]) Get() (T, error) {
	return ir.result, ir.err
}

func (ir IntermediateResult[T]) Must() T {
	if ir.err != nil {
		panic(ir.err)
	}
	return ir.result
}

func NewHookService() IntermediateResult[HookService] {
	plugins, err := load(dir)
	if err != nil {
		return IntermediateResult[HookService]{err: err}
	}
	return IntermediateResult[HookService]{
		result: HookService{plugins: plugins},
	}
}

type HookService struct {
	plugins []*plugin.Plugin
}

func (s HookService) OnInit() error {
	for _, p := range s.plugins {
		h, err := p.Lookup("OnInit")
		if err != nil {
			continue
		}
		hook, ok := h.(func() error)
		if !ok {
			return errors.New("invalid OnInit plugin hook installed")
		}
		if err := hook(); err != nil {
			return err
		}
	}
	return nil
}

func (s HookService) BeforeRequestPrepared(ctx context.Context, cmd *cobra.Command, ex *exchange.Exchange) error {
	for _, p := range s.plugins {
		h, err := p.Lookup("BeforeRequestPrepared")
		if err != nil {
			continue
		}
		hook, ok := h.(func(context.Context, *cobra.Command, *exchange.Exchange) error)
		if !ok {
			return errors.New("invalid BeforeRequestPrepared plugin hook installed")
		}
		if err := hook(ctx, cmd, ex); err != nil {
			return err
		}
	}
	return nil
}

func (s HookService) BeforeRequestDispatched(ctx context.Context, cmd *cobra.Command, ex *exchange.Exchange, r *http.Request) error {
	for _, p := range s.plugins {
		h, err := p.Lookup("BeforeRequestDispatched")
		if err != nil {
			continue
		}
		hook, ok := h.(func(context.Context, *cobra.Command, *exchange.Exchange, *http.Request) error)
		if !ok {
			return errors.New("invalid BeforeRequestDispatched plugin hook installed")
		}
		if err := hook(ctx, cmd, ex, r); err != nil {
			return err
		}
	}
	return nil
}

func (s HookService) OnResponse(ctx context.Context, cmd *cobra.Command, ex *exchange.Exchange, r *http.Response) error {
	for _, p := range s.plugins {
		h, err := p.Lookup("OnResponse")
		if err != nil {
			continue
		}
		hook, ok := h.(func(context.Context, *cobra.Command, *exchange.Exchange, *http.Response) error)
		if !ok {
			return errors.New("invalid OnResponse plugin hook installed")
		}
		if err := hook(ctx, cmd, ex, r); err != nil {
			return err
		}
	}
	return nil
}
