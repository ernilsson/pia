package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/ernilsson/pia/environment"
	"github.com/ernilsson/pia/exchange"
	"github.com/spf13/cobra"
)

var target string

var do = &cobra.Command{
	Use:   "do",
	Short: "sends a request according to the specified configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		if verbose {
			log.Default().Printf("using working directory: %s\n", wd)
		}
		env, err := environment.Load(wd)
		if err != nil {
			return err
		}
		if verbose {
			log.Default().Printf("loaded environment: %+v", env)
		}
		f, err := os.Open(target)
		if err != nil {
			return err
		}
		defer f.Close()

		scn := bufio.NewScanner(f)
		buf := bytes.NewBufferString("")
		for scn.Scan() {
			line := scn.Text()
			if env.RequiresSubstitution(line) {
				line, err = env.Substitute(line)
				if err != nil {
					return err
				}
			}
			_, err := buf.WriteString(line + "\n")
			if err != nil {
				return err
			}
		}
		if verbose {
			log.Default().Println("performed yaml configuration substitution")
		}
		ex, err := exchange.LoadReader(buf)
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
	root.AddCommand(do)
}
