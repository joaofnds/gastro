package driver

import (
	"astro/test/req"
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type API struct {
	baseURL string
}

func NewAPI(baseURL string) *API {
	return &API{baseURL}
}

func (a API) List(token string) (*http.Response, error) {
	return req.Get(
		a.baseURL+"/habits",
		map[string]string{"Authorization": token},
	)
}

func (a API) Create(token, name string) (*http.Response, error) {
	return req.Post(
		a.baseURL+"/habits?name="+name,
		map[string]string{"Content-Type": "application/json", "Authorization": token},
		new(bytes.Buffer),
	)
}

func (a API) Update(token, habitID, name string) (*http.Response, error) {
	return req.Patch(
		a.baseURL+"/habits/"+habitID,
		map[string]string{"Content-Type": "application/json", "Authorization": token},
		strings.NewReader(fmt.Sprintf(`{"name":%q}`, name)),
	)
}

func (a API) Get(token, id string) (*http.Response, error) {
	return req.Get(
		a.baseURL+"/habits/"+id,
		map[string]string{"Authorization": token},
	)
}

func (a API) Delete(token, id string) (*http.Response, error) {
	return req.Delete(
		a.baseURL+"/habits/"+id,
		map[string]string{"Authorization": token},
	)
}

func (a API) AddActivity(token, id, desc string, date time.Time) (*http.Response, error) {
	return req.Post(
		a.baseURL+"/habits/"+id,
		map[string]string{"Authorization": token, "Content-Type": "application/json"},
		strings.NewReader(fmt.Sprintf(`{"description":%q,"date":%q}`, desc, date.UTC().Format(time.RFC3339))),
	)
}

func (a API) UpdateActivity(token, habitID, activityID, desc string) (*http.Response, error) {
	return req.Patch(
		a.baseURL+"/habits/"+habitID+"/"+activityID,
		map[string]string{"Authorization": token, "Content-Type": "application/json"},
		strings.NewReader(fmt.Sprintf(`{"description":%q}`, desc)),
	)
}

func (a API) DeleteActivity(token, habitID, activityID string) (*http.Response, error) {
	return req.Delete(
		a.baseURL+"/habits/"+habitID+"/"+activityID,
		map[string]string{"Authorization": token},
	)
}

func (a API) CreateGroup(token, name string) (*http.Response, error) {
	return req.Post(
		a.baseURL+"/groups",
		map[string]string{"Authorization": token, "Content-Type": "application/json"},
		strings.NewReader(fmt.Sprintf(`{"name":%q}`, name)),
	)
}

func (a API) AddToGroup(token, habitID, groupID string) (*http.Response, error) {
	return req.Post(
		a.baseURL+"/groups/"+groupID+"/"+habitID,
		map[string]string{"Authorization": token},
		new(bytes.Buffer),
	)
}

func (a API) RemoveFromGroup(token, habitID, groupID string) (*http.Response, error) {
	return req.Delete(
		a.baseURL+"/groups/"+groupID+"/"+habitID,
		map[string]string{"Authorization": token},
	)
}

func (a API) DeleteGroup(token, groupID string) (*http.Response, error) {
	return req.Delete(
		a.baseURL+"/groups/"+groupID,
		map[string]string{"Authorization": token},
	)
}

func (a API) GroupsAndHabits(token string) (*http.Response, error) {
	return req.Get(
		a.baseURL+"/groups",
		map[string]string{"Authorization": token},
	)
}

func (a API) CreateToken() (*http.Response, error) {
	return req.Post(
		a.baseURL+"/token",
		map[string]string{"Content-Type": "application/json"},
		new(bytes.Buffer),
	)
}

func (a API) TestToken(token string) (*http.Response, error) {
	return req.Get(
		a.baseURL+"/tokentest",
		map[string]string{"Authorization": token},
	)
}
