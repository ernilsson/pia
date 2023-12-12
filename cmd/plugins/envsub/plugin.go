package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/ernilsson/pia/exchange"
	"github.com/ernilsson/pia/hook"
	"io"
	"net/http"
	"os"
	"regexp"
)

var BeforeRequestDispatched = before()

func before() hook.BeforeRequestDispatchedHook {
	return func(ex *exchange.Exchange, r *http.Request) error {
		d, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		d, err = substituteLines(d)
		if err != nil {
			return err
		}
		r.Body = io.NopCloser(bytes.NewReader(d))
		return nil
	}
}

func substituteLines(d []byte) ([]byte, error) {
	scn := bufio.NewScanner(bytes.NewReader(d))
	buf := new(bytes.Buffer)
	for scn.Scan() {
		line, err := substituteLine(scn.Text())
		if err != nil {
			return nil, err
		}
		if _, err := fmt.Fprintln(buf, line); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func substituteLine(line string) (string, error) {
	regx := regexp.MustCompile("\\$\\{env\\..+?}")
	if !regx.MatchString(line) {
		return line, nil
	}
	regx = regexp.MustCompile("\\$\\{env\\.(.+?)}")
	matches := regx.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		val := os.Getenv(match[1])
		regx = regexp.MustCompile(fmt.Sprintf("\\$\\{env.%s\\}", match[1]))
		line = regx.ReplaceAllString(line, fmt.Sprintf("%v", val))
	}
	return line, nil
}
