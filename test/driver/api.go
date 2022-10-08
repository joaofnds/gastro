package driver

import (
	"astro/http/util"
	"bytes"
	"net/http"
)

type API struct {
	baseURL string
}

func NewAPI(baseURL string) *API {
	return &API{baseURL}
}

func (a API) List(token string) (*http.Response, error) {
	return util.Get(
		a.baseURL+"/habits",
		map[string]string{"Authorization": token},
	)
}

func (a API) Create(token, name string) (*http.Response, error) {
	return util.Post(
		a.baseURL+"/habits?name="+name,
		map[string]string{"Content-Type": "application/json", "Authorization": token},
		new(bytes.Buffer),
	)
}

func (a API) Get(token, id string) (*http.Response, error) {
	return util.Get(
		a.baseURL+"/habits/"+id,
		map[string]string{"Authorization": token},
	)
}

func (a API) Delete(token, id string) (*http.Response, error) {
	return util.Delete(
		a.baseURL+"/habits/"+id,
		map[string]string{"Authorization": token},
	)
}

func (a API) AddActivity(token, id string) (*http.Response, error) {
	return util.Post(
		a.baseURL+"/habits/"+id,
		map[string]string{"Authorization": token},
		new(bytes.Buffer),
	)
}

func (a API) CreateToken() (*http.Response, error) {
	return util.Post(
		a.baseURL+"/token",
		map[string]string{"Content-Type": "application/json"},
		new(bytes.Buffer),
	)
}

func (a API) TestToken(token string) (*http.Response, error) {
	return util.Get(
		a.baseURL+"/tokentest",
		map[string]string{"Authorization": token},
	)
}
