package exchange

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

var Nil = Exchange{}

func Load(conf string, vars map[string]any) (Exchange, error) {
	var ex Exchange
	if err := yaml.Unmarshal([]byte(conf), &ex); err != nil {
		return Nil, err
	}
	for key, val := range ex.Request.Body.Variables {
		filled := val
		v, ok := vars[key]
		if !ok {
			if filled.Default != nil {
				filled.Value = filled.Default
			} else {
				return Nil, fmt.Errorf("no value provided for variable '%s'", key)
			}
		} else {
			filled.Value = v
		}
		delete(ex.Request.Body.Variables, key)
		ex.Request.Body.Variables[key] = filled
	}
	if ex.Request.URL == "" {
		return Nil, errors.New("no url provided")
	}
	if ex.Request.Method == "" {
		return Nil, errors.New("no method provided")
	}
	for key, val := range ex.Request.Body.Variables {
		if val.Required && val.Value == nil {
			return Nil, fmt.Errorf("no value provided for required variable '%s'", key)
		}
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

func (rc RequestConfiguration) emptyBody() bool {
	return rc.Body.Template == ""
}

type BodyConfiguration struct {
	Template  string              `yaml:"template"`
	Variables map[string]Variable `yaml:"variables"`
}

type Variable struct {
	Required bool `yaml:"required"`
	Value    any
	Default  any `yaml:"default"`
}

func (e Exchange) Do() (*http.Response, error) {
	url, err := url.Parse(e.Request.URL)
	if err != nil {
		return nil, err
	}
	req := http.Request{
		URL:    url,
		Method: e.Request.Method,
		Header: e.headers(),
	}
	body, err := e.body()
	if err != nil {
		return nil, err
	}
	req.Body = io.NopCloser(bytes.NewBuffer(body))
	return http.DefaultClient.Do(&req)
}

func (e Exchange) headers() http.Header {
	headers := make(http.Header)
	for key, val := range e.Request.Headers {
		headers[key] = append(headers[key], val)
	}
	return headers
}

func (e Exchange) body() ([]byte, error) {
	if e.Request.emptyBody() {
		return nil, nil
	}
	tmpl, err := os.Open(e.Request.Body.Template)
	if err != nil {
		return nil, err
	}
	defer tmpl.Close()
	b, err := io.ReadAll(tmpl)
	if err != nil {
		return nil, err
	}
	return e.substituteTmpl(b)
}

func (e Exchange) substituteTmpl(tmpl []byte) ([]byte, error) {
	buf := bytes.NewBufferString("")
	scn := bufio.NewScanner(bytes.NewReader(tmpl))
	for scn.Scan() {
		line := scn.Text()
		var err error
		line, err = e.substituteLine(line)
		if err != nil {
			return nil, err
		}
		_, err = fmt.Fprintln(buf, line)
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func (e Exchange) substituteLine(line string) (string, error) {
	if !e.requiresSubstitution(line) {
		return line, nil
	}
	reg := regexp.MustCompile("\\$\\{var\\.(.+?)\\}{1}")
	targets := reg.FindAllStringSubmatch(line, -1)
	m := make(map[string]any)
	for _, target := range targets {
		res, ok := e.Request.Body.Variables[target[1]]
		if !ok {
			return "", fmt.Errorf("unkown variable '%s'", target)
		}
		m[target[1]] = res.Value
	}
	for key, val := range m {
		exp := regexp.MustCompile(fmt.Sprintf("\\$\\{var.%s\\}", key))
		line = exp.ReplaceAllString(line, fmt.Sprintf("%v", val))
	}
	return line, nil
}

func (e Exchange) requiresSubstitution(line string) bool {
	reg := regexp.MustCompile("\\$\\{var\\..+?\\}")
	return reg.Match([]byte(line))
}
