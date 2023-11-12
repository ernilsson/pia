package environment

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

func Bootstrap(dir string) error {
	f, err := os.Create(fmt.Sprintf("%s/env.json", dir))
	if err != nil {
		return err
	}
	defer f.Close()

	env := Environment(map[string]any{
		"scheme": "https",
		"port":   443,
	})
	m, err := json.Marshal(&env)
	if err != nil {
		return err
	}
	cnt, err := f.Write(m)
	if err != nil {
		return err
	}
	if cnt < len(m) {
		return errors.New("failed to write boilerplate environment file")
	}
	return nil
}

func Load(r io.Reader) (Environment, error) {
	m, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var env Environment
	if err := json.Unmarshal(m, &env); err != nil {
		return nil, err
	}
	return env, nil
}

type Environment map[string]any

func (e Environment) SubstituteReader(r io.Reader) (string, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return e.SubstituteLines(string(b))
}

func (e Environment) SubstituteLines(r string) (string, error) {
	buf := bytes.NewBufferString("")
	scn := bufio.NewScanner(bytes.NewBufferString(r))
	for scn.Scan() {
		line, err := e.substitute(scn.Text())
		if err != nil {
			return "", err
		}
		_, err = fmt.Fprintln(buf, line)
		if err != nil {
			return "", err
		}
	}
	return buf.String(), nil
}

func (e Environment) substitute(line string) (string, error) {
	if !e.templated(line) {
		return line, nil
	}
	reg := regexp.MustCompile("\\$\\{env\\.(.+?)\\}{1}")
	targets := reg.FindAllStringSubmatch(line, -1)
	flattened, err := e.flatten(targets)
	if err != nil {
		return "", err
	}
	for key, val := range flattened {
		exp := regexp.MustCompile(fmt.Sprintf("\\$\\{env.%s\\}", key))
		line = exp.ReplaceAllString(line, fmt.Sprintf("%v", val))
	}
	return line, nil
}

func (e Environment) templated(line string) bool {
	reg := regexp.MustCompile("\\$\\{env\\..+?\\}")
	return reg.Match([]byte(line))
}

func (e Environment) flatten(targets [][]string) (map[string]any, error) {
	m := make(map[string]any)
	for _, target := range targets {
		res, err := e.resolve(target[1])
		if err != nil {
			return nil, err
		}
		m[target[1]] = res
	}
	return m, nil
}

func (e Environment) resolve(key string) (any, error) {
	segments := strings.Split(key, ".")
	leaf := segments[len(segments)-1]
	segments = segments[:len(segments)-1]
	v := e
	for _, segment := range segments {
		_, ok := v[segment]
		if !ok {
			return nil, fmt.Errorf("referenced non-existent environment variable '%s'", key)
		}
		_, ok = v[segment].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("referenced non-existent environment variable '%s'", key)
		}
		v = v[segment].(map[string]any)
	}
	val, ok := v[leaf]
	if !ok {
		return nil, fmt.Errorf("references non-existent environment variable '%s'", key)
	}
	return val, nil
}
