package plugin

import (
	"context"
	"errors"
	"github.com/ernilsson/pia/exchange"
	"github.com/spf13/cobra"
	"plugin"
)

func NewAugmenterService() (*AugmenterService, error) {
	plugins, err := load(dir)
	if err != nil {
		return nil, err
	}
	return &AugmenterService{plugins: plugins}, nil
}

type AugmenterService struct {
	plugins []*plugin.Plugin
}

func (s *AugmenterService) ExchangePreProcessors(ctx context.Context, cmd *cobra.Command) ([]exchange.PreProcessor, error) {
	processors := make([]exchange.PreProcessor, 0)
	for _, p := range s.plugins {
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

func (s *AugmenterService) BodyPreProcessors(ctx context.Context, cmd *cobra.Command) ([]exchange.PreProcessor, error) {
	processors := make([]exchange.PreProcessor, 0)
	for _, p := range s.plugins {
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
