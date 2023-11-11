package environment

import (
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

func Load(dir string) (Environment, error) {
	f, err := os.Open(fmt.Sprintf("%s/env.json", dir))
	if err != nil {
		return nil, err
	}
	m, err := io.ReadAll(f)
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

func (e Environment) RequiresSubstitution(line string) bool {
	reg := regexp.MustCompile("\\$\\{env\\..+?\\}")
	return reg.Match([]byte(line))
}

func (e Environment) Substitute(line string) (string, error) {
	if !e.RequiresSubstitution(line) {
		return line, errors.New("line does not require environment substitution")
	}
	reg := regexp.MustCompile("\\$\\{env\\.(.+?)\\}{1}")
	targets := reg.FindAllStringSubmatch(line, -1)
	m := make(map[string]any)
	for _, target := range targets {
		res, err := e.resolve(target[1])
		if err != nil {
			return "", err
		}
		m[target[1]] = res
	}
	for key, val := range m {
		exp := regexp.MustCompile(fmt.Sprintf("\\$\\{env.%s\\}", key))
		line = exp.ReplaceAllString(line, fmt.Sprintf("%v", val))
	}
	return line, nil
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
	return v[leaf], nil
}
