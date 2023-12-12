package hook

import (
	"errors"
	"github.com/ernilsson/pia/exchange"
	"net/http"
)

const pluginDir = "/usr/local/bin/pia-plug/"

type BeforeRequestPreparedHook func(*exchange.Exchange) error

type BeforeRequestDispatchedHook func(*exchange.Exchange, *http.Request) error

type OnResponseHook func(*exchange.Exchange, *http.Response) error

func BeforeRequestPrepared(ex *exchange.Exchange) error {
	plugins, err := load(pluginDir)
	if err != nil {
		return err
	}
	for _, p := range plugins {
		h, err := p.Lookup("BeforeRequestPrepared")
		if err != nil {
			continue
		}
		hook, ok := h.(BeforeRequestPreparedHook)
		if !ok {
			return errors.New("invalid BeforeRequestPrepared hook installed")
		}
		if err := hook(ex); err != nil {
			return err
		}
	}
	return nil
}

func BeforeRequestDispatched(ex *exchange.Exchange, r *http.Request) error {
	plugins, err := load(pluginDir)
	if err != nil {
		return err
	}
	for _, p := range plugins {
		h, err := p.Lookup("BeforeRequestDispatched")
		if err != nil {
			continue
		}
		hook, ok := h.(BeforeRequestDispatchedHook)
		if !ok {
			return errors.New("invalid BeforeRequestDispatched hook installed")
		}
		if err := hook(ex, r); err != nil {
			return err
		}
	}
	return nil
}

func OnResponse(ex *exchange.Exchange, r *http.Response) error {
	plugins, err := load(pluginDir)
	if err != nil {
		return err
	}
	for _, p := range plugins {
		h, err := p.Lookup("OnResponse")
		if err != nil {
			continue
		}
		hook, ok := h.(OnResponseHook)
		if !ok {
			return errors.New("invalid OnResponse hook installed")
		}
		if err := hook(ex, r); err != nil {
			return err
		}
	}
	return nil
}
