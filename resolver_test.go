package pia_test

import (
	"github.com/ernilsson/pia"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDelegatingKeyResolver_Resolve(t *testing.T) {
	tests := []struct {
		key       string
		delegates map[string]pia.KeyResolver
		value     string
		err       error
	}{
		{
			key: "session:id_token",
			delegates: map[string]pia.KeyResolver{
				"session": pia.MapResolver{
					"id_token": "abc",
				},
			},
			value: "abc",
		},
		{
			key: "session:id_token:for_me",
			delegates: map[string]pia.KeyResolver{
				"session": pia.MapResolver{
					"id_token:for_me": "abc",
				},
			},
			value: "abc",
		},
		{
			key: "session:not_found",
			delegates: map[string]pia.KeyResolver{
				"session": pia.MapResolver{
					"id_token": "abc",
				},
			},
			err: pia.ErrKeyNotFound,
		},
		{
			key: "env:not_found",
			delegates: map[string]pia.KeyResolver{
				"session": pia.MapResolver{
					"id_token": "abc",
				},
			},
			err: pia.ErrKeyNotFound,
		},
		{
			key: "env:id_token",
			delegates: map[string]pia.KeyResolver{
				"session": pia.MapResolver{
					"id_token": "abc",
				},
			},
			err: pia.ErrKeyNotFound,
		},
	}
	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			value, err := pia.DelegatingKeyResolver{
				Delegates: test.delegates,
			}.Resolve(test.key)
			assert.ErrorIs(t, err, test.err)
			assert.Equal(t, test.value, value)
		})
	}
}
