package exchange

import (
	"fmt"
	"io"
	"os"
)

type ProviderFunc func() ([]byte, error)

func FileProvider(path string) ProviderFunc {
	return func() ([]byte, error) {
		f, err := os.OpenFile(path, os.O_RDONLY, os.ModeAppend)
		if err != nil {
			return nil, err
		}
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				panic(err)
			}
		}(f)
		return io.ReadAll(f)
	}
}

func DiscoveringFileProvider(path string) (ProviderFunc, error) {
	mutators := []filePathMutator{exact(), extension("yml"), extension("yaml")}
	for _, mut := range mutators {
		mutated := mut(path)
		info, err := os.Stat(mutated)
		if err == nil && !info.IsDir() {
			return FileProvider(mutated), nil
		}
	}
	return nil, os.ErrNotExist
}

type filePathMutator func(fp string) string

func exact() filePathMutator {
	return func(fp string) string {
		return fp
	}
}

func extension(ext string) filePathMutator {
	return func(fp string) string {
		return fmt.Sprintf("%s.%s", fp, ext)
	}
}
