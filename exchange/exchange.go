package exchange

import (
	"io"
	"net/http"
	"net/url"

	"gopkg.in/yaml.v3"
)

var Nil = Exchange{}

func LoadReader(r io.Reader) (Exchange, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return Nil, err
	}
	var ex Exchange
	if err := yaml.Unmarshal(b, &ex); err != nil {
		return Nil, err
	}
	return ex, nil
}

type Exchange struct {
	Version     string               `yaml:"version"`
	Name        string               `yaml:"name"`
	Description string               `yaml:"description"`
	Request     RequestConfiguration `yaml:"request"`
}

type RequestConfiguration struct {
	Method  string            `yaml:"method"`
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers"`
	Body    BodyConfiguration `yaml:"body"`
}

type BodyConfiguration struct {
	Template  string              `yaml:"template"`
	Variables map[string]Variable `yaml:"variables"`
}

type Variable struct {
	Required bool `yaml:"required"`
	Default  any  `yaml:"default"`
}

func (e Exchange) Do() (*http.Response, error) {
	url, err := url.Parse(e.Request.URL)
	if err != nil {
		return nil, err
	}
	headers := make(http.Header)
	for key, val := range e.Request.Headers {
		headers[key] = append(headers[key], val)
	}
	req := http.Request{
		URL:    url,
		Method: e.Request.Method,
		Header: headers,
	}
	return http.DefaultClient.Do(&req)
}
