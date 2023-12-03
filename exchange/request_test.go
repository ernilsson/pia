package exchange

import (
	"errors"
	"net/http"
	"testing"
)

func Test_NewRequest_CopiesHeaderToRequest(t *testing.T) {
	config := RequestConfiguration{
		Headers: map[string]string{
			"Accept":          "application/json",
			"X-Custom-Header": "test-header",
		},
	}
	req, err := NewRequest(config)
	if err != nil {
		t.Errorf("got unexpected error: %s", err)
		return
	}
	if req.Header["Accept"][0] != "application/json" {
		t.Errorf("expected 'application/json' in 'Accept' header but got: %s", req.Header["Accept"][0])
		return
	}
	if req.Header["X-Custom-Header"][0] != "test-header" {
		t.Errorf("expected 'test-header' in 'X-Custom-Header' header but got: %s", req.Header["X-Custom-Header"][0])
		return
	}
}

func Test_NewRequest_SetsRequestMethod(t *testing.T) {
	config := RequestConfiguration{
		Method: "POST",
	}
	req, err := NewRequest(config)
	if err != nil {
		t.Errorf("got unexpected error: %s", err)
		return
	}
	if req.Method != "POST" {
		t.Errorf("expected 'POST' method but got: %s", req.Method)
		return
	}

	config.Method = "PATCH"
	req, err = NewRequest(config)
	if err != nil {
		t.Errorf("got unexpected error: %s", err)
		return
	}
	if req.Method != "PATCH" {
		t.Errorf("expected 'PATCH' method but got: %s", req.Method)
		return
	}
}

func Test_NewRequest_SetsURL(t *testing.T) {
	config := RequestConfiguration{
		URL: "https://test.com/endpoint",
	}
	req, err := NewRequest(config)
	if err != nil {
		t.Errorf("got unexpected error: %s", err)
		return
	}
	if req.URL.String() != "https://test.com/endpoint" {
		t.Errorf("expected 'https://test.com/endpoint' URL but got: %s", req.URL)
		return
	}
}

func Test_NewRequest_GivenErroneousOption_ReturnsError(t *testing.T) {
	config := RequestConfiguration{}
	opt := NewRequestOption(func(rc RequestConfiguration, req *http.Request) error {
		return errors.New("mocked error")
	})
	_, err := NewRequest(config, opt)
	if err == nil {
		t.Errorf("expected error but got nil")
		return
	}
}
