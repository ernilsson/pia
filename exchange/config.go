package exchange

import (
	"bufio"
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path"
	"regexp"
	"strings"
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
	Variables    VariableSet
}

type VariableSet map[string]string

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

func (v VariableSet) SubstituteLine(line string) (string, error) {
	regx := regexp.MustCompile("\\$\\{var\\..+?}")
	if !regx.MatchString(line) {
		return line, nil
	}
	regx = regexp.MustCompile("\\$\\{var\\.(.+?)}")
	matches := regx.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		val, err := v.resolve(match[1])
		if err != nil {
			return "", err
		}
		regx = regexp.MustCompile(fmt.Sprintf("\\$\\{var.%s\\}", match[1]))
		line = regx.ReplaceAllString(line, fmt.Sprintf("%v", val))
	}
	return line, nil
}

func (v VariableSet) resolve(key string) (string, error) {
	segments := strings.Split(key, ":")
	if len(segments) > 2 {
		return "", fmt.Errorf("malformed key for variable '%s'", key)
	}
	val, ok := v[segments[0]]
	if !ok && len(segments) < 2 {
		return "", fmt.Errorf("variable '%s' not set and has no default value", key)
	}
	if !ok {
		return segments[1], nil
	}
	return val, nil
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
