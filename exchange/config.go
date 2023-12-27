package exchange

import (
	"bytes"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
	"path"
)

type SubstitutionSource interface {
	SubstituteLines(data []byte) ([]byte, error)
}

type PreProcessor func(raw []byte) ([]byte, error)

func AdaptSubstitution(sub SubstitutionSource) PreProcessor {
	return func(raw []byte) ([]byte, error) {
		return sub.SubstituteLines(raw)
	}
}

type NewExchangeOption func(ex *Exchange) error

func ConfigurationPreProcessor(pp PreProcessor) NewExchangeOption {
	return func(ex *Exchange) error {
		data, err := yaml.Marshal(ex)
		if err != nil {
			return err
		}
		data, err = pp(data)
		return yaml.Unmarshal(data, ex)
	}
}

func NewExchange(provider ProviderFunc, opts ...NewExchangeOption) (Exchange, error) {
	data, err := provider()
	if err != nil {
		return Exchange{}, err
	}
	var ex Exchange
	if err := yaml.Unmarshal(data, &ex); err != nil {
		return Exchange{}, err
	}
	for _, opt := range opts {
		err = opt(&ex)
		if err != nil {
			return Exchange{}, err
		}
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

type BuildRequestOption func(Exchange, *http.Request) error

func BodyPreProcessor(processor PreProcessor) BuildRequestOption {
	return func(ex Exchange, req *http.Request) error {
		body, err := ex.RequestBody()
		if err != nil {
			return err
		}
		if body == nil {
			return nil
		}
		body, err = processor(body)
		if err != nil {
			return err
		}
		req.Body = io.NopCloser(bytes.NewReader(body))
		return nil
	}
}

func (ex Exchange) BuildRequest(opts ...BuildRequestOption) (*http.Request, error) {
	req, err := http.NewRequest(ex.Request.Method, ex.Request.URL, nil)
	if err != nil {
		return nil, err
	}
	for key, value := range ex.Request.Headers {
		req.Header[key] = []string{value}
	}
	for _, opt := range opts {
		if err := opt(ex, req); err != nil {
			return nil, err
		}
	}
	return req, nil
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
