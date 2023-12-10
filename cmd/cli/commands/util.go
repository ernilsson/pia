package commands

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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

func DiscoverExchangeFile(fp string) (string, error) {
	opts := []FilePathMutator{Extension("yml"), Extension("yaml")}
	attempt, err := DiscoverFile(fp, opts...)
	if err != nil {
		return DiscoverFile(filepath.Join(fp, "config"), opts...)
	}
	return attempt, nil
}

type FilePathMutator func(filepath string) string

func Exact() FilePathMutator {
	return func(filepath string) string {
		return filepath
	}
}

func Extension(extension string) FilePathMutator {
	return func(filepath string) string {
		return fmt.Sprintf("%s.%s", filepath, extension)
	}
}

func DiscoverFile(filepath string, opts ...FilePathMutator) (string, error) {
	opts = append(opts, Exact())
	for _, mut := range opts {
		mutated := mut(filepath)
		info, err := os.Stat(mutated)
		if err == nil && !info.IsDir() {
			return mutated, nil
		}
	}
	return "", os.ErrNotExist
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
