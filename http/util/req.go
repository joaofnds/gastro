package util

import (
	"bytes"
	"io"
	"net/http"
)

type Headers map[string]string

func Req(method string, url string, headers Headers, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	return http.DefaultClient.Do(req)
}

func Get(url string, headers Headers) (*http.Response, error) {
	return Req(http.MethodGet, url, headers, new(bytes.Buffer))
}

func Post(url string, headers Headers, body io.Reader) (*http.Response, error) {
	return Req(http.MethodPost, url, headers, body)
}

func Delete(url string, headers Headers) (*http.Response, error) {
	return Req(http.MethodDelete, url, headers, new(bytes.Buffer))
}
