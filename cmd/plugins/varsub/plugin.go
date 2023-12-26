package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/ernilsson/pia/cmd/cli/commands"
	"github.com/ernilsson/pia/exchange"
	"github.com/spf13/cobra"
	"regexp"
	"strings"
)

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

func ExchangePreProcessorFactory(ctx context.Context, cmd *cobra.Command) (exchange.PreProcessor, error) {
	pairs, err := cmd.Flags().GetStringSlice("variable")
	if err != nil {
		return nil, err
	}
	vars := make(VariableSet)
	for _, pair := range pairs {
		key, val, ok := strings.Cut(pair, "=")
		if !ok {
			return nil, fmt.Errorf("invalid pair '%s'", pair)
		}
		vars[key] = val
	}
	return func(raw []byte) ([]byte, error) {
		return vars.SubstituteLines(raw)
	}, nil
}

func BodyPreProcessorFactory(ctx context.Context, cmd *cobra.Command) (exchange.PreProcessor, error) {
	pairs, err := cmd.Flags().GetStringSlice("variable")
	if err != nil {
		return nil, err
	}
	vars := make(VariableSet)
	for _, pair := range pairs {
		key, val, ok := strings.Cut(pair, "=")
		if !ok {
			return nil, fmt.Errorf("invalid pair '%s'", pair)
		}
		vars[key] = val
	}
	return func(raw []byte) ([]byte, error) {
		return vars.SubstituteLines(raw)
	}, nil
}

func OnInit() error {
	commands.Prepare.Flags().StringSliceP("variable", "v", nil, "sets a variable")
	return nil
}
