package command

import (
	"context"
	"github.com/ernilsson/pia/exchange"
	"github.com/ernilsson/pia/plugin"
	"github.com/ernilsson/pia/profile"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

func NewRequestFactory() RequestFactory {
	augmenters := plugin.NewAugmenterService().Must()
	store := profile.Must(profile.NewFileStore(must(os.Getwd())))
	return RequestFactory{
		augmenters:    augmenters,
		activeProfile: store.Must(store.LoadActive()),
	}
}

type RequestFactory struct {
	augmenters    plugin.AugmenterService
	activeProfile profile.Profile
}

func (f RequestFactory) Create(ctx context.Context, cmd *cobra.Command, ex exchange.Exchange) (*http.Request, error) {
	processors, err := f.augmenters.BodyPreProcessors(ctx, cmd)
	if err != nil {
		return nil, err
	}
	processors = append(processors, exchange.AdaptSubstitution(f.activeProfile))
	opts := make([]exchange.BuildRequestOption, 0)
	for _, processor := range processors {
		opts = append(opts, exchange.BodyPreProcessor(processor))
	}
	return ex.BuildRequest(opts...)
}

func NewExchangeFactory() ExchangeFactory {
	augmenters := plugin.NewAugmenterService().Must()
	store := profile.Must(profile.NewFileStore(must(os.Getwd())))
	return ExchangeFactory{
		augmenters:    augmenters,
		activeProfile: store.Must(store.LoadActive()),
	}
}

type ExchangeFactory struct {
	augmenters    plugin.AugmenterService
	activeProfile profile.Profile
}

func (e ExchangeFactory) Create(ctx context.Context, cmd *cobra.Command, file string) (exchange.Exchange, error) {
	provider, err := exchange.DiscoveringFileProvider(file)
	if err != nil {
		return exchange.Exchange{}, err
	}
	processors := e.augmenters.Must(e.augmenters.ExchangePreProcessors(ctx, cmd))
	processors = append(processors, exchange.AdaptSubstitution(e.activeProfile))
	opts := make([]exchange.NewExchangeOption, 0)
	for _, processor := range processors {
		opts = append(opts, exchange.ConfigurationPreProcessor(processor))
	}
	return exchange.NewExchange(provider, opts...)
}

var Do = &cobra.Command{
	Use:        "do",
	Short:      "executes a request and writes the response to stdout",
	Args:       cobra.ExactArgs(1),
	ArgAliases: []string{"exchange configuration file"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		hooks := plugin.NewHookService().Must()

		ex, err := NewExchangeFactory().Create(ctx, cmd, args[0])
		if err != nil {
			return err
		}
		if err := hooks.BeforeRequestPrepared(ctx, cmd, &ex); err != nil {
			return err
		}
		req, err := NewRequestFactory().Create(ctx, cmd, ex)
		if err != nil {
			return err
		}
		if err := hooks.BeforeRequestDispatched(ctx, cmd, &ex, req); err != nil {
			return err
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		if err := hooks.OnResponse(ctx, cmd, &ex, res); err != nil {
			return err
		}
		return WriteResponse(os.Stdout, res)
	},
}

var Prepare = &cobra.Command{
	Use:        "prepare",
	Aliases:    []string{"Prepare"},
	Short:      "prepares a request without executing it and writes the result to stdout",
	Args:       cobra.ExactArgs(1),
	ArgAliases: []string{"exchange configuration file"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		hooks := plugin.NewHookService().Must()
		ex, err := NewExchangeFactory().Create(ctx, cmd, args[0])
		if err != nil {
			return err
		}
		if err := hooks.BeforeRequestPrepared(ctx, cmd, &ex); err != nil {
			return err
		}
		req, err := NewRequestFactory().Create(ctx, cmd, ex)
		if err != nil {
			return err
		}
		return WriteRequest(os.Stdout, req)
	},
}

func init() {
	Root.AddCommand(Do)
	Root.AddCommand(Prepare)
}
