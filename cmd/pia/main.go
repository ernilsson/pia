package main

import (
	"bufio"
	"flag"
	"github.com/ernilsson/pia/cmd/pia/internal/tui"
	"log"
	"os"
	"strings"
)

func main() {
	flag.Parse()
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	props := make(map[string]string)
	if flag.NArg() > 0 {
		props, err = properties(flag.Arg(0))
		if err != nil {
			log.Fatalln(err)
		}
	}
	if err := tui.Run(wd, props); err != nil {
		log.Fatalln(err)
	}
}

func properties(path string) (map[string]string, error) {
	props := make(map[string]string)
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	scn := bufio.NewScanner(strings.NewReader(string(src)))
	for scn.Scan() {
		line := scn.Text()
		segments := strings.SplitN(line, "=", 2)
		if len(segments) != 2 {
			continue
		}
		props[segments[0]] = strings.TrimSpace(
			strings.Trim(segments[1], "\""),
		)
	}
	return props, nil
}
