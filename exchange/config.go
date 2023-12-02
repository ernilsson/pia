package exchange

import (
	"github.com/ernilsson/pia/profile"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

type ProviderFunc func() ([]byte, error)

func FileProvider(path string) ProviderFunc {
	return func() ([]byte, error) {
		f, err := os.OpenFile(path, os.O_RDONLY, os.ModeAppend)
		if err != nil {
			return nil, err
		}
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				panic(err)
			}
		}(f)
		return io.ReadAll(f)
	}
}

func GetExchange(p profile.Profile, pf ProviderFunc) (Exchange, error) {
	data, err := pf()
	if err != nil {
		return Exchange{}, err
	}
	data, err = p.SubstituteLines(data)
	if err != nil {
		return Exchange{}, err
	}
	var ex Exchange
	if err := yaml.Unmarshal(data, &ex); err != nil {
		return Exchange{}, err
	}
	return ex, nil
}

type Exchange struct {
	Request RequestConfiguration `yaml:"request"`
}

type RequestConfiguration struct {
	Method  string            `yaml:"method"`
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers"`
	Body    BodyConfiguration `yaml:"body"`
}

type BodyConfiguration struct {
	TemplateFile string `yaml:"template"`
}

func (bc BodyConfiguration) Template() ([]byte, error) {
	if bc.empty() {
		return nil, nil
	}
	f, err := os.OpenFile(bc.TemplateFile, os.O_RDONLY, os.ModeAppend)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (bc BodyConfiguration) empty() bool {
	return bc.TemplateFile == ""
}
