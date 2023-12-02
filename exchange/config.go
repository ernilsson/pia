package exchange

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/ernilsson/pia/profile"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"regexp"
	"strings"
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

func GetExchange(pf ProviderFunc, p profile.Profile, vars VariableSet) (Exchange, error) {
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
	ex.Request.Body.Variables = vars
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
	Variables    VariableSet
}

type VariableSet map[string]any

func (v VariableSet) SubstituteLine(line string) (string, error) {
	regx := regexp.MustCompile("\\$\\{var\\..+?}")
	if !regx.MatchString(line) {
		return line, nil
	}
	regx = regexp.MustCompile("\\$\\{var\\.(.+?)}")
	matches := regx.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		// TODO: Refactor default value extraction (possibly make value resolution its own func)
		var defaultVal any
		defaults := strings.Split(match[1], ":")
		if len(defaults) == 2 {
			defaultVal = defaults[1]
		}
		val, ok := v[match[1]]
		if !ok && defaultVal == nil {
			return "", errors.New("variable '" + match[1] + "' not set and has no default value")
		}
		if !ok {
			val = defaultVal
		}
		regx = regexp.MustCompile(fmt.Sprintf("\\$\\{var.%s\\}", match[1]))
		line = regx.ReplaceAllString(line, fmt.Sprintf("%v", val))
	}
	return line, nil
}

func (v VariableSet) SubstituteLines(data []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	scn := bufio.NewScanner(bytes.NewReader(data))
	for scn.Scan() {
		line, err := v.SubstituteLine(scn.Text())
		if err != nil {
			return nil, err
		}
		if _, err := fmt.Fprintln(buf, line); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
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
