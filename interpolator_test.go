package pia_test

import (
	"errors"
	"github.com/ernilsson/pia"
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
)

func TestSubstitutingReader_Read(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		tests := []struct {
			name     string
			resolver map[string]string
			input    string

			readBufferSize int
			expected       string
		}{
			{
				name: "entire stream fitting within buffer",
				resolver: map[string]string{
					"env.username": "pia",
					"env.password": "P@$$W0RD",
				},
				input: `
				{
					"username": "${env.username}",
					"password": "${env.password}"
				}
				`,
				expected: `
				{
					"username": "pia",
					"password": "P@$$W0RD"
				}
				`,
				readBufferSize: 512,
			},
			{
				name: "chopped substitution point",
				resolver: map[string]string{
					"env.username": "pia",
				},
				input:          "${env.username}",
				expected:       "pia",
				readBufferSize: 4,
			},
			{
				name: "ending with dollar sign but missing brace",
				resolver: map[string]string{
					"env.username": "pia",
				},
				input:          "   ${env.username}",
				expected:       "   pia",
				readBufferSize: 4,
			},
			{
				name: "ending with dollar sign without requiring substitution",
				resolver: map[string]string{
					"env.username": "pia",
				},
				input:          "   $",
				expected:       "   $",
				readBufferSize: 4,
			},
			{
				name: "non-terminated substitution point",
				resolver: map[string]string{
					"env.username": "pia",
				},
				input:          "${env.username",
				expected:       "${env.username",
				readBufferSize: 4,
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				r := pia.WrapReader(
					pia.MapResolver(test.resolver),
					strings.NewReader(test.input),
				)

				var actual string
				for {
					buf := make([]byte, test.readBufferSize)
					n, err := r.Read(buf)
					if errors.Is(err, io.EOF) {
						break
					}
					assert.Nil(t, err)
					actual += string(buf[:n])
				}

				assert.Equal(t, test.expected, actual)
			})
		}
	})

	t.Run("sad path", func(t *testing.T) {
		tests := []struct {
			name           string
			resolver       map[string]string
			input          string
			err            error
			readBufferSize int
		}{
			{
				name:           "missing key",
				input:          "${env.username}",
				resolver:       map[string]string{},
				err:            pia.ErrKeyNotFound,
				readBufferSize: 512,
			},
			{
				name:           "insufficient destination length",
				input:          "${env.username}",
				err:            pia.ErrInsufficientDestinationLength,
				readBufferSize: 1,
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				r := pia.WrapReader(
					pia.MapResolver(test.resolver),
					strings.NewReader(test.input),
				)
				for {
					buf := make([]byte, test.readBufferSize)
					_, err := r.Read(buf)
					if errors.Is(err, io.EOF) {
						break
					}
					if err != nil {
						assert.True(t, errors.Is(err, test.err))
						break
					}
				}
			})
		}
	})
}
