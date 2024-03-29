package driver

import (
	"encoding/json"
	"io"
	"net/http"
)

type params struct {
	into   any
	status int
	req    func() (*http.Response, error)
}

func makeJSONRequest(p params) error {
	res, err := p.req()
	if err != nil {
		return err
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != p.status {
		return RequestFailure{Status: res.StatusCode, Body: string(b)}
	}

	if p.into == nil {
		return nil
	}

	return json.Unmarshal(b, p.into)
}

type RequestFailure struct {
	Status int
	Body   string
}

func (e RequestFailure) Error() string {
	return e.Body
}
