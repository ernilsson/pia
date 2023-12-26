package main

import (
	"github.com/ernilsson/pia/cmd/cli/command"
	"github.com/ernilsson/pia/hook"
)

func main() {
	if err := hook.OnInit(); err != nil {
		panic(err)
	}
	command.Execute()
}
