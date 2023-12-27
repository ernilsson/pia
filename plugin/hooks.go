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

func NewHookService() (*HookService, error) {
	plugins, err := load(dir)
	if err != nil {
		return nil, err
	}
	return &HookService{
		plugins: plugins,
	}, nil
}

type HookService struct {
	plugins []*plugin.Plugin
}

func (s *HookService) OnInit() error {
	for _, p := range s.plugins {
		h, err := p.Lookup("OnInit")
		if err != nil {
			continue
		}
		hook, ok := h.(func() error)
		if !ok {
			return errors.New("invalid OnInit plugin installed")
		}
		if err := hook(); err != nil {
			return err
		}
	}
	return nil
}

func (s *HookService) BeforeRequestPrepared(ctx context.Context, cmd *cobra.Command, ex *exchange.Exchange) error {
	for _, p := range s.plugins {
		h, err := p.Lookup("BeforeRequestPrepared")
		if err != nil {
			continue
		}
		hook, ok := h.(func(context.Context, *cobra.Command, *exchange.Exchange) error)
		if !ok {
			return errors.New("invalid BeforeRequestPrepared plugin installed")
		}
		if err := hook(ctx, cmd, ex); err != nil {
			return err
		}
	}
	return nil
}

func (s *HookService) BeforeRequestDispatched(ctx context.Context, cmd *cobra.Command, ex *exchange.Exchange, r *http.Request) error {
	for _, p := range s.plugins {
		h, err := p.Lookup("BeforeRequestDispatched")
		if err != nil {
			continue
		}
		hook, ok := h.(func(context.Context, *cobra.Command, *exchange.Exchange, *http.Request) error)
		if !ok {
			return errors.New("invalid BeforeRequestDispatched plugin installed")
		}
		if err := hook(ctx, cmd, ex, r); err != nil {
			return err
		}
	}
	return nil
}

func (s *HookService) OnResponse(ctx context.Context, cmd *cobra.Command, ex *exchange.Exchange, r *http.Response) error {
	for _, p := range s.plugins {
		h, err := p.Lookup("OnResponse")
		if err != nil {
			continue
		}
		hook, ok := h.(func(context.Context, *cobra.Command, *exchange.Exchange, *http.Response) error)
		if !ok {
			return errors.New("invalid OnResponse plugin installed")
		}
		if err := hook(ctx, cmd, ex, r); err != nil {
			return err
		}
	}
	return nil
}

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
			return errors.New("invalid OnInit plugin installed")
		}
		if err := hook(); err != nil {
			return err
		}
	}
	return nil
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
			return errors.New("invalid BeforeRequestPrepared plugin installed")
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
			return errors.New("invalid BeforeRequestDispatched plugin installed")
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
			return errors.New("invalid OnResponse plugin installed")
		}
		if err := hook(ctx, cmd, ex, r); err != nil {
			return err
		}
	}
	return nil
}
