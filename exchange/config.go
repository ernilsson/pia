package exchange

import (
	"bufio"
	"bytes"
	"fmt"
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
		defer must(f.Close())
		return io.ReadAll(f)
	}
}

func GetExchange(p profile.Profile, pf ProviderFunc) (Exchange, error) {
	data, err := pf()
	if err != nil {
		return Exchange{}, err
	}
	buf := new(bytes.Buffer)
	scn := bufio.NewScanner(bytes.NewReader(data))
	for scn.Scan() {
		line, err := p.SubstituteLine(scn.Text())
		if err != nil {
			return Exchange{}, err
		}
		if _, err := fmt.Fprintln(buf, line); err != nil {
			return Exchange{}, err
		}
	}
	var ex Exchange
	if err := yaml.Unmarshal(buf.Bytes(), &ex); err != nil {
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
	defer must(f.Close())
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (bc BodyConfiguration) empty() bool {
	return bc.TemplateFile == ""
}
