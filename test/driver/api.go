package driver

import (
	"fmt"
	"net/http"
	"strings"
)

type API struct{}

func NewAPI() *API {
	return &API{}
}

func (a API) List() (*http.Response, error) {
	return http.Get("http://localhost:3000/habits")
}

func (a API) Create(name string) (*http.Response, error) {
	url := fmt.Sprintf("http://localhost:3000/habits?name=%s", name)
	return http.Post(url, "application/text", strings.NewReader(""))
}

func (a API) Get(name string) (*http.Response, error) {
	url := fmt.Sprintf("http://localhost:3000/habits/%s", name)
	return http.Get(url)
}

func (a API) Delete(name string) (*http.Response, error) {
	url := fmt.Sprintf("http://localhost:3000/habits/%s", name)
	req, err := http.NewRequest(http.MethodDelete, url, strings.NewReader(""))
	if err != nil {
		return nil, err
	}

	return http.DefaultClient.Do(req)
}

func (a API) AddActivity(name string) (*http.Response, error) {
	url := fmt.Sprintf("http://localhost:3000/habits/%s", name)
	return http.Post(url, "application/json", strings.NewReader(""))
}
