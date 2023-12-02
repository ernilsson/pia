package exchange

import (
	"bytes"
	"github.com/ernilsson/pia/profile"
	"net/http"
)

func NewRequest(p profile.Profile, configuration RequestConfiguration) (*http.Request, error) {
	headers := make(http.Header)
	for key, value := range configuration.Headers {
		headers[key] = []string{value}
	}
	body, err := configuration.Body.Template()
	if err != nil {
		return nil, err
	}
	body, err = p.SubstituteLines(body)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest(configuration.Method, configuration.URL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	request.Header = headers
	return request, nil
}
