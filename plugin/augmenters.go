package plugin

import (
	"context"
	"errors"
	"github.com/ernilsson/pia/exchange"
	"github.com/spf13/cobra"
	"plugin"
)

func NewAugmenterService() IntermediateResult[AugmenterService] {
	plugins, err := load(dir)
	if err != nil {
		return IntermediateResult[AugmenterService]{err: err}
	}
	return IntermediateResult[AugmenterService]{
		result: AugmenterService{plugins: plugins},
	}
}

type AugmenterService struct {
	plugins []*plugin.Plugin
}

func (s *AugmenterService) Must(processors []exchange.PreProcessor, err error) []exchange.PreProcessor {
	if err != nil {
		panic(err)
	}
	return processors
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
