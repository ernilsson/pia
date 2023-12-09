package commands

import (
	"errors"
	"fmt"
	"github.com/ernilsson/pia/exchange"
	"github.com/ernilsson/pia/profile"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

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
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	file := path.Join(wd, input)
	info, ok, err := FileExists(file)
	if err != nil {
		return "", err
	}
	if ok && info.IsDir() {
		return DiscoverExchangeFile(path.Join(input, "config"))
	} else if ok {
		return file, nil
	}

	file = path.Join(wd, fmt.Sprintf("%s.yml", input))
	info, ok, err = FileExists(file)
	if err != nil {
		return "", err
	}
	if ok {
		return file, nil
	}

	file = path.Join(wd, fmt.Sprintf("%s.yaml", input))
	info, ok, err = FileExists(file)
	if err != nil {
		return "", err
	}
	if ok {
		return file, nil
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

func PrepareRequest(filepath string, vars map[string]string) (*http.Request, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	store := profile.NewFileStore(wd)
	prof, err := store.LoadActive()
	if err != nil {
		return nil, err
	}
	ex, err := exchange.GetExchange(
		exchange.FileProvider(filepath),
		exchange.TemplatedConfiguration(prof),
	)
	if err != nil {
		return nil, err
	}
	ex.ConfigRoot = path.Dir(filepath)
	return exchange.NewRequest(ex, exchange.TemplatedBody(prof, exchange.VariableSet(vars)))
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
