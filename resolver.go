package pia

import (
	"fmt"
	"os"
	"strings"
)

type DelegatingKeyResolver struct {
	Delegates map[string]KeyResolver
}

func (d DelegatingKeyResolver) Resolve(k string) (string, error) {
	sections := strings.SplitN(k, ":", 2)
	delegate, ok := d.Delegates[sections[0]]
	if !ok {
		return "", fmt.Errorf("%w: %s is not a valid delegate", ErrKeyNotFound, sections[0])
	}
	return delegate.Resolve(sections[1])
}

type FallbackResolverDecorator struct {
	Delegate KeyResolver
}

func (f FallbackResolverDecorator) Resolve(k string) (string, error) {
	key, fb, ok := strings.Cut(k, "|")
	if v, err := f.Delegate.Resolve(key); err != nil {
		if ok {
			return fb, nil
		}
		return "", err
	} else {
		return v, nil
	}
}

type EnvironmentResolver struct{}

func (e EnvironmentResolver) Resolve(k string) (string, error) {
	value := os.Getenv(k)
	if value == "" {
		return "", fmt.Errorf("%w: %s is not in environment", ErrKeyNotFound, k)
	}
	return value, nil
}

// KeyResolver is a simple abstraction of a key-value store for string values.
type KeyResolver interface {
	// Resolve takes a key and returns a value from the underlying store. An error value may be returned for any
	// erroneous reason but for the case where a key cannot be resolved it is expected to return [pia.ErrKeyNotFound].
	Resolve(k string) (string, error)
}

// MapResolver is the simplest possible implementation of [pia.KeyResolver] using an underlying map to facilitate
// storage and retrieval.
type MapResolver map[string]string

// Resolve implements the [pia.KeyResolver] interface.
func (m MapResolver) Resolve(k string) (string, error) {
	v, ok := m[k]
	if !ok {
		return "", fmt.Errorf("failed to resolve key '%s': %w", k, ErrKeyNotFound)
	}
	return v, nil
}
