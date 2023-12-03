package plug

import (
	"github.com/ernilsson/pia/exchange"
	"net/http"
)

type RequestHook interface {
	OnRequest(ex exchange.Exchange, req *http.Request) error
}

func LoadRequestHooks(dir string) ([]RequestHook, error) {
	plugs, err := load(dir)
	if err != nil {
		return nil, err
	}
	hooks := make([]RequestHook, 0)
	for _, plug := range plugs {
		hook, err := plug.Lookup("RequestHook")
		if err != nil {
			return nil, err
		}
		if h, ok := hook.(RequestHook); ok {
			hooks = append(hooks, h)
		}
	}
	return hooks, nil
}
