package exchange

import (
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path"
)

type PreProcessor func(raw []byte) ([]byte, error)

func TemplatedConfiguration(sub ...SubstitutionSource) PreProcessor {
	return func(raw []byte) ([]byte, error) {
		var err error
		for _, s := range sub {
			raw, err = s.SubstituteLines(raw)
			if err != nil {
				return nil, err
			}
		}
		return raw, nil
	}
}

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

func GetExchange(provider ProviderFunc, processors ...PreProcessor) (Exchange, error) {
	data, err := provider()
	if err != nil {
		return Exchange{}, err
	}
	for _, processor := range processors {
		data, err = processor(data)
		if err != nil {
			return Exchange{}, err
		}
	}
	var ex Exchange
	if err := yaml.Unmarshal(data, &ex); err != nil {
		return Exchange{}, err
	}
	return ex, nil
}

type Exchange struct {
	ConfigRoot string

	Version string               `yaml:"version"`
	Request RequestConfiguration `yaml:"request"`
}

func (ex Exchange) InConfigRoot(filename string) string {
	return path.Join(ex.ConfigRoot, filename)
}

type RequestConfiguration struct {
	Method  string            `yaml:"method"`
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers"`
	Body    BodyConfiguration `yaml:"body"`
}

type BodyConfiguration struct {
	TemplateFile string `yaml:"file"`
}

func (ex Exchange) RequestBody() ([]byte, error) {
	if ex.Request.Body.empty() {
		return nil, nil
	}
	f, err := os.OpenFile(ex.InConfigRoot(ex.Request.Body.TemplateFile), os.O_RDONLY, os.ModeAppend)
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
