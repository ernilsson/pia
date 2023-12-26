package exchange

import (
	"bytes"
	"io"
	"net/http"
)

type SubstitutionSource interface {
	SubstituteLines(data []byte) ([]byte, error)
}

type NewRequestOption func(ex Exchange, req *http.Request) error

func SubstitutionPreProcessor(src SubstitutionSource) PreProcessor {
	return func(raw []byte) ([]byte, error) {
		return src.SubstituteLines(raw)
	}
}

func PreProcessedBody(processors ...PreProcessor) NewRequestOption {
	return func(ex Exchange, req *http.Request) error {
		body, err := ex.RequestBody()
		if err != nil {
			return err
		}
		for _, processor := range processors {
			body, err = processor(body)
			if err != nil {
				return err
			}
		}
		req.Body = io.NopCloser(bytes.NewReader(body))
		return nil
	}
}

func NewRequest(ex Exchange, opts ...NewRequestOption) (*http.Request, error) {
	req, err := http.NewRequest(ex.Request.Method, ex.Request.URL, nil)
	if err != nil {
		return nil, err
	}
	for key, value := range ex.Request.Headers {
		req.Header[key] = []string{value}
	}
	for _, opt := range opts {
		if err := opt(ex, req); err != nil {
			return nil, err
		}
	}
	return req, nil
}
