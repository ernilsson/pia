package plug

import (
	"github.com/ernilsson/pia/exchange"
	"net/http"
)

type ResponseHook interface {
	OnResponse(ex exchange.Exchange, res *http.Response) error
}

func LoadResponseHooks(dir string) ([]RequestHook, error) {
	plugs, err := load(dir)
	if err != nil {
		return nil, err
	}
	hooks := make([]RequestHook, 0)
	for _, plug := range plugs {
		hook, err := plug.Lookup("ResponseHook")
		if err != nil {
			return nil, err
		}
		if h, ok := hook.(RequestHook); ok {
			hooks = append(hooks, h)
		}
	}
	return hooks, nil
}
