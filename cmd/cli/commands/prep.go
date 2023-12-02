package commands

import (
	"fmt"
	"github.com/ernilsson/pia/exchange"
	"github.com/ernilsson/pia/profile"
	"github.com/spf13/cobra"
	"io"
	"os"
)

var prep = &cobra.Command{
	Use:     "prepare",
	Aliases: []string{"prep"},
	Short:   "prepares a request without executing it and writes the result to stdout",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prof, err := profile.UnmarshalActive()
		if err != nil {
			return err
		}
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		target := args[0]
		ex, err := exchange.GetExchange(prof, exchange.FileProvider(fmt.Sprintf("%s/%s", wd, target)))
		if err != nil {
			return err
		}
		req, err := exchange.NewRequest(prof, ex.Request)
		if err != nil {
			return err
		}
		fmt.Printf("URL: %s\n", req.URL)
		fmt.Printf("Method: %s\n", req.Method)
		for key, v := range req.Header {
			fmt.Printf("%s: %s\n", key, v[0])
		}
		fmt.Println()
		parsed, err := io.ReadAll(req.Body)
		if err != nil {
			return err
		}
		fmt.Println(string(parsed))
		return nil
	},
}

func init() {
	root.AddCommand(prep)
}
