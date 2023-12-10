package commands

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

func MustParse(vars map[string]string, err error) map[string]string {
	if err != nil {
		panic(err)
	}
	return vars
}

func ParseKeyValues(pairs []string) (map[string]string, error) {
	kv := make(map[string]string)
	for _, pair := range pairs {
		key, val, ok := strings.Cut(pair, "=")
		if !ok {
			return nil, fmt.Errorf("invalid pair '%s'", pair)
		}
		kv[key] = val
	}
	return kv, nil
}

func DiscoverExchangeFile(input string) (string, error) {
	info, ok, err := FileExists(input)
	if err != nil {
		return "", err
	}
	if ok && info.IsDir() {
		return DiscoverExchangeFile(path.Join(input, "config"))
	} else if ok {
		return input, nil
	}

	yml := fmt.Sprintf("%s.yml", input)
	info, ok, err = FileExists(yml)
	if err != nil {
		return "", err
	}
	if ok {
		return yml, nil
	}

	yaml := fmt.Sprintf("%s.yaml", input)
	info, ok, err = FileExists(yaml)
	if err != nil {
		return "", err
	}
	if ok {
		return yaml, nil
	}
	return "", errors.New("config file not found")
}

func FileExists(filepath string) (os.FileInfo, bool, error) {
	info, err := os.Stat(filepath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, false, err
	} else if err != nil && errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	return info, true, nil
}

func WriteRequest(w io.Writer, req *http.Request) error {
	if _, err := fmt.Fprintf(w, "URL: %s\n", req.URL); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Method: %s\n", req.Method); err != nil {
		return err
	}
	for key, v := range req.Header {
		if _, err := fmt.Fprintf(w, "%s: %s\n", key, v[0]); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(body))
	return err
}

func WriteResponse(w io.Writer, res *http.Response) error {
	for key, v := range res.Header {
		if _, err := fmt.Fprintf(w, "%s: %s\n", key, v[0]); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(body))
	return err
}
