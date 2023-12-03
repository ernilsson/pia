package exchange

import (
	"bytes"
	"io"
	"net/http"
)

type SubstitutionSource interface {
	SubstituteLines(data []byte) ([]byte, error)
}

type NewRequestOption func(rc RequestConfiguration, req *http.Request) error

func TemplatedBody(src ...SubstitutionSource) NewRequestOption {
	return func(rc RequestConfiguration, req *http.Request) error {
		body, err := rc.Body.Template()
		if err != nil {
			return err
		}
		for _, sub := range src {
			body, err = sub.SubstituteLines(body)
			if err != nil {
				return err
			}
		}
		req.Body = io.NopCloser(bytes.NewReader(body))
		return nil
	}
}

func NewRequest(rc RequestConfiguration, opts ...NewRequestOption) (*http.Request, error) {
	req, err := http.NewRequest(rc.Method, rc.URL, nil)
	if err != nil {
		return nil, err
	}
	for key, value := range rc.Headers {
		req.Header[key] = []string{value}
	}
	for _, opt := range opts {
		if err := opt(rc, req); err != nil {
			return nil, err
		}
	}
	return req, nil
}
