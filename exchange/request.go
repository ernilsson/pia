package exchange

import (
	"bufio"
	"bytes"
	"fmt"
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
	buf := new(bytes.Buffer)
	scn := bufio.NewScanner(bytes.NewReader(body))
	for scn.Scan() {
		line, err := p.SubstituteLine(scn.Text())
		if err != nil {
			return nil, err
		}
		if _, err := fmt.Fprintln(buf, line); err != nil {
			return nil, err
		}
	}
	request, err := http.NewRequest(configuration.Method, configuration.URL, buf)
	if err != nil {
		return nil, err
	}
	request.Header = headers
	return request, nil
}
