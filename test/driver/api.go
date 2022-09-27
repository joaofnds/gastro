package driver

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type API struct{}

func NewAPI() *API {
	return &API{}
}

type Headers = map[string]string

func (a API) List(token string) (*http.Response, error) {
	return get("http://localhost:3000/habits", map[string]string{"Authorization": token})
}

func (a API) Create(token, name string) (*http.Response, error) {
	return post(
		"http://localhost:3000/habits?name="+name,
		Headers{"Content-Type": "application/text", "Authorization": token},
		&bytes.Buffer{},
	)
}

func (a API) Get(token, name string) (*http.Response, error) {
	return get(
		"http://localhost:3000/habits/"+name,
		map[string]string{"Authorization": token},
	)
}

func (a API) Delete(token, name string) (*http.Response, error) {
	url := fmt.Sprintf("http://localhost:3000/habits/%s", name)
	req, err := http.NewRequest(http.MethodDelete, url, strings.NewReader(""))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token)

	return http.DefaultClient.Do(req)
}

func (a API) AddActivity(token, name string) (*http.Response, error) {
	return post(
		"http://localhost:3000/habits/"+name,
		map[string]string{"Authorization": token},
		&bytes.Buffer{},
	)
}

func (a API) CreateToken() (*http.Response, error) {
	return http.Post("http://localhost:3000/token", "application/text", strings.NewReader(""))
}

func (a API) TestToken(token []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, "http://localhost:3000/tokentest", strings.NewReader(""))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", string(token))
	return http.DefaultClient.Do(req)
}

func get(url string, headers Headers) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, new(bytes.Buffer))
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	return http.DefaultClient.Do(req)
}

func post(url string, headers Headers, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, &bytes.Buffer{})
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	return http.DefaultClient.Do(req)
}
