package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ernilsson/pia/environment"
	"github.com/ernilsson/pia/exchange"
	"github.com/spf13/cobra"
)

var (
	target    string
	variables []string
)

var do = &cobra.Command{
	Use:   "do",
	Short: "sends a request according to the specified configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		conf, err := os.Open(fmt.Sprintf("%s/env.json", wd))
		if err != nil {
			return err
		}
		env, err := environment.Load(conf)
		if err != nil {
			return err
		}

		f, err := os.Open(target)
		if err != nil {
			return err
		}
		defer f.Close()
		parsed, err := env.SubstituteReader(f)
		if err != nil {
			return err
		}
		vars := make(map[string]any)
		for _, v := range variables {
			keyval := strings.SplitN(v, "=", 2)
			vars[keyval[0]] = keyval[1]
		}
		ex, err := exchange.Load(parsed, vars)
		if err != nil {
			return err
		}
		resp, err := ex.Do()
		if err != nil {
			return err
		}
		if body, err := io.ReadAll(resp.Body); err != nil {
			return err
		} else {
			fmt.Println(string(body))
		}
		return nil
	},
}

func init() {
	do.Flags().StringVarP(&target, "target", "t", "", "target configuration to use")
	do.Flags().StringArrayVar(&variables, "var", nil, "overloaded variable values")
	root.AddCommand(do)
}
