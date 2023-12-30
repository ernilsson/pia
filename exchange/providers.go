package exchange

import (
	"fmt"
	"io"
	"os"
	"sync"
)

type ProviderFunc func() ([]byte, string, error)

type Cache interface {
	Populated() bool
	Get() ([]byte, string)
	Set([]byte, string)
}

type SimpleCache struct {
	lock    *sync.RWMutex
	data    []byte
	locator string
}

func (s SimpleCache) Populated() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.data != nil
}

func (s SimpleCache) Get() ([]byte, string) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.data, s.locator
}

func (s SimpleCache) Set(data []byte, locator string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.data = data
	s.locator = locator
}

func NewSimpleCache() *SimpleCache {
	return &SimpleCache{
		lock:    &sync.RWMutex{},
		data:    nil,
		locator: "",
	}
}

func CachedProvider(provider ProviderFunc, cache Cache) ProviderFunc {
	return func() ([]byte, string, error) {
		if cache.Populated() {
			data, locator := cache.Get()
			return data, locator, nil
		}
		data, locator, err := provider()
		if err != nil {
			return nil, "", err
		}
		cache.Set(data, locator)
		return data, locator, nil
	}
}

func PreProcessedProvider(provider ProviderFunc, pp PreProcessor) ProviderFunc {
	return func() ([]byte, string, error) {
		data, locator, err := provider()
		if err != nil {
			return nil, "", err
		}
		data, err = pp(data)
		if err != nil {
			return nil, "", err
		}
		return data, locator, nil
	}
}

func DiscoveringFileProvider(path string) (ProviderFunc, error) {
	mutators := []mutator{exact(), extension("yml"), extension("yaml")}
	for _, mut := range mutators {
		mutated := mut(path)
		info, err := os.Stat(mutated)
		if err == nil && !info.IsDir() {
			return FileProvider(mutated), nil
		}
	}
	return nil, os.ErrNotExist
}

type mutator func(fp string) string

func exact() mutator {
	return func(fp string) string {
		return fp
	}
}

func extension(ext string) mutator {
	return func(fp string) string {
		return fmt.Sprintf("%s.%s", fp, ext)
	}
}

func FileProvider(path string) ProviderFunc {
	return func() ([]byte, string, error) {
		f, err := os.OpenFile(path, os.O_RDONLY, os.ModeAppend)
		if err != nil {
			return nil, "", err
		}
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				panic(err)
			}
		}(f)
		data, err := io.ReadAll(f)
		if err != nil {
			return nil, "", err
		}
		return data, path, nil
	}
}
