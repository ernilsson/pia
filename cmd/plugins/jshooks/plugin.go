package main

import (
	"context"
	"github.com/ernilsson/pia/cmd/plugins/jshooks/internal"
	"github.com/ernilsson/pia/exchange"
	"github.com/ernilsson/pia/profile"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
	"path"
)

type HookConfiguration struct {
	Hooks Hooks `yaml:"hooks"`
}

type Hooks struct {
	BeforeRequestPrepared   string `yaml:"before_request_prepared"`
	BeforeRequestDispatched string `yaml:"before_request_dispatched"`
	OnResponse              string `yaml:"on_response"`
}

func loadScript(name string) ([]byte, error) {
	script, err := os.OpenFile(name, os.O_RDONLY, os.ModeAppend)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(script)
}

func BeforeRequestPrepared(ctx context.Context, cmd *cobra.Command, ex *exchange.Exchange) error {
	data, loc, err := ex.Source()
	if err != nil {
		return err
	}
	var config HookConfiguration
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}
	if config.Hooks.BeforeRequestPrepared == "" {
		return nil
	}
	src, err := loadScript(path.Join(path.Dir(loc), config.Hooks.BeforeRequestPrepared))
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	vm, err := internal.New(ex, internal.ProfileSetter(profile.Must(profile.NewFileStore(wd))))
	_, err = vm.Run(src)
	return err
}

func BeforeRequestDispatched(ctx context.Context, cmd *cobra.Command, ex *exchange.Exchange, req *http.Request) error {
	data, loc, err := ex.Source()
	if err != nil {
		return err
	}
	var config HookConfiguration
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}
	if config.Hooks.BeforeRequestDispatched == "" {
		return nil
	}
	src, err := loadScript(path.Join(path.Dir(loc), config.Hooks.BeforeRequestDispatched))
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	vm, err := internal.New(
		ex,
		internal.ProfileSetter(profile.Must(profile.NewFileStore(wd))),
		internal.Value("request", req),
	)
	_, err = vm.Run(src)
	return err
}

func OnResponse(ctx context.Context, cmd *cobra.Command, ex *exchange.Exchange, res *http.Response) error {
	data, loc, err := ex.Source()
	if err != nil {
		return err
	}
	var config HookConfiguration
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}
	if config.Hooks.OnResponse == "" {
		return nil
	}
	src, err := loadScript(path.Join(path.Dir(loc), config.Hooks.OnResponse))
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	vm, err := internal.New(
		ex,
		internal.ProfileSetter(profile.Must(profile.NewFileStore(wd))),
		internal.Value("response", res),
	)
	_, err = vm.Run(src)
	return err
}
