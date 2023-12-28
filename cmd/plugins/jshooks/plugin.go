package main

import (
	"context"
	"errors"
	"github.com/ernilsson/pia/exchange"
	"github.com/robertkrimen/otto"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path"
	"path/filepath"
)

type HooksConfiguration struct {
	BeforeRequestPrepared string `yaml:"before_request_prepared"`
}

func BeforeRequestPrepared(ctx context.Context, cmd *cobra.Command, ex *exchange.Exchange) error {
	file, err := os.OpenFile(filepath.Join(ex.ConfigRoot, "hooks.yml"), os.O_RDONLY, os.ModeAppend)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	} else if err != nil && errors.Is(err, os.ErrNotExist) {
		return nil
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	var config HooksConfiguration
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}
	if config.BeforeRequestPrepared == "" {
		return nil
	}
	script, err := os.OpenFile(path.Join(ex.ConfigRoot, config.BeforeRequestPrepared), os.O_RDONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	src, err := io.ReadAll(script)
	if err != nil {
		return err
	}
	vm := otto.New()
	if err := vm.Set("exchange", ex); err != nil {
		return err
	}
	_, err = vm.Run(src)
	return err
}
