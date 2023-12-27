package main

import (
	"github.com/ernilsson/pia/cmd/cli/command"
	"github.com/ernilsson/pia/plugin"
)

func main() {
	if err := plugin.NewHookService().Must().OnInit(); err != nil {
		panic(err)
	}
	command.Execute()
}
