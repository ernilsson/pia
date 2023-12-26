package main

import (
	"github.com/ernilsson/pia/cmd/cli/commands"
	"github.com/ernilsson/pia/hook"
)

func main() {
	if err := hook.OnInit(); err != nil {
		panic(err)
	}
	commands.Execute()
}
