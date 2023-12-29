package internal

import (
	"github.com/ernilsson/pia/exchange"
	"github.com/ernilsson/pia/profile"
	"github.com/robertkrimen/otto"
)

type Option func(*otto.Otto) error

// Value sets a value on the JavaScript VM instance.
func Value(name string, val any) Option {
	return func(o *otto.Otto) error {
		return o.Set(name, val)
	}
}

// Func currently does the exact same thing as Value and only exists as a semantic difference to make intention clearer.
// Will ideally add a runtime check to make sure that value passed in is indeed a function.
func Func(name string, val any) Option {
	return Value(name, val)
}

func ProfileSetter(store profile.Store) Option {
	return Func("setProfileValue", func(call otto.FunctionCall) {
		active, err := store.LoadActive()
		if err != nil {
			return
		}
		key, err := call.Argument(0).ToString()
		if err != nil {
			return
		}
		val, err := call.Argument(1).ToString()
		if err != nil {
			return
		}
		active[key] = val
		_ = store.Save(active)
		return
	})
}

func New(ex *exchange.Exchange, opts ...Option) (*otto.Otto, error) {
	opts = append(opts, Value("exchange", ex))
	vm := otto.New()
	for _, opt := range opts {
		if err := opt(vm); err != nil {
			return nil, err
		}
	}
	return vm, nil
}
