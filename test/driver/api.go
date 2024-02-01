package driver

import (
	"astro/test/matchers"
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

func (a API) MustList(token string) *http.Response {
	return matchers.Must2(a.List(token))
}

func (a API) Create(token, name string) (*http.Response, error) {
	return req.Post(
		a.baseURL+"/habits?name="+name,
		map[string]string{"Content-Type": "application/json", "Authorization": token},
		new(bytes.Buffer),
	)
}

func (a API) MustCreate(token, name string) *http.Response {
	return matchers.Must2(a.Create(token, name))
}

func (a API) Update(token, habitID, name string) (*http.Response, error) {
	return req.Patch(
		a.baseURL+"/habits/"+habitID,
		map[string]string{"Content-Type": "application/json", "Authorization": token},
		strings.NewReader(fmt.Sprintf(`{"name":%q}`, name)),
	)
}

func (a API) MustUpdate(token, habitID, name string) *http.Response {
	return matchers.Must2(a.Update(token, habitID, name))
}

func (a API) Get(token, id string) (*http.Response, error) {
	return req.Get(
		a.baseURL+"/habits/"+id,
		map[string]string{"Authorization": token},
	)
}

func (a API) MustGet(token, id string) *http.Response {
	return matchers.Must2(a.Get(token, id))
}

func (a API) Delete(token, id string) (*http.Response, error) {
	return req.Delete(
		a.baseURL+"/habits/"+id,
		map[string]string{"Authorization": token},
	)
}

func (a API) MustDelete(token, id string) *http.Response {
	return matchers.Must2(a.Delete(token, id))
}

func (a API) AddActivity(token, id, desc string, date time.Time) (*http.Response, error) {
	return req.Post(
		a.baseURL+"/habits/"+id,
		map[string]string{"Authorization": token, "Content-Type": "application/json"},
		strings.NewReader(fmt.Sprintf(`{"description":%q,"date":%q}`, desc, date.UTC().Format(time.RFC3339))),
	)
}

func (a API) MustAddActivity(token, id, desc string, date time.Time) *http.Response {
	return matchers.Must2(a.AddActivity(token, id, desc, date))
}

func (a API) UpdateActivity(token, habitID, activityID, desc string) (*http.Response, error) {
	return req.Patch(
		a.baseURL+"/habits/"+habitID+"/"+activityID,
		map[string]string{"Authorization": token, "Content-Type": "application/json"},
		strings.NewReader(fmt.Sprintf(`{"description":%q}`, desc)),
	)
}

func (a API) MustUpdateActivity(token, habitID, activityID, desc string) *http.Response {
	return matchers.Must2(a.UpdateActivity(token, habitID, activityID, desc))
}

func (a API) DeleteActivity(token, habitID, activityID string) (*http.Response, error) {
	return req.Delete(
		a.baseURL+"/habits/"+habitID+"/"+activityID,
		map[string]string{"Authorization": token},
	)
}

func (a API) MustDeleteActivity(token, habitID, activityID string) *http.Response {
	return matchers.Must2(a.DeleteActivity(token, habitID, activityID))
}

func (a API) CreateGroup(token, name string) (*http.Response, error) {
	return req.Post(
		a.baseURL+"/groups",
		map[string]string{"Authorization": token, "Content-Type": "application/json"},
		strings.NewReader(fmt.Sprintf(`{"name":%q}`, name)),
	)
}

func (a API) MustCreateGroup(token, name string) *http.Response {
	return matchers.Must2(a.CreateGroup(token, name))
}

func (a API) AddToGroup(token, habitID, groupID string) (*http.Response, error) {
	return req.Post(
		a.baseURL+"/groups/"+groupID+"/"+habitID,
		map[string]string{"Authorization": token},
		new(bytes.Buffer),
	)
}

func (a API) MustAddToGroup(token, habitID, groupID string) *http.Response {
	return matchers.Must2(a.AddToGroup(token, habitID, groupID))
}

func (a API) RemoveFromGroup(token, habitID, groupID string) (*http.Response, error) {
	return req.Delete(
		a.baseURL+"/groups/"+groupID+"/"+habitID,
		map[string]string{"Authorization": token},
	)
}

func (a API) MustRemoveFromGroup(token, habitID, groupID string) *http.Response {
	return matchers.Must2(a.RemoveFromGroup(token, habitID, groupID))
}

func (a API) DeleteGroup(token, groupID string) (*http.Response, error) {
	return req.Delete(
		a.baseURL+"/groups/"+groupID,
		map[string]string{"Authorization": token},
	)
}

func (a API) MustDeleteGroup(token, groupID string) *http.Response {
	return matchers.Must2(a.DeleteGroup(token, groupID))
}

func (a API) GroupsAndHabits(token string) (*http.Response, error) {
	return req.Get(
		a.baseURL+"/groups",
		map[string]string{"Authorization": token},
	)
}

func (a API) MustGroupsAndHabits(token string) *http.Response {
	return matchers.Must2(a.GroupsAndHabits(token))
}

func (a API) CreateToken() (*http.Response, error) {
	return req.Post(
		a.baseURL+"/token",
		map[string]string{"Content-Type": "application/json"},
		new(bytes.Buffer),
	)
}

func (a API) MustCreateToken() *http.Response {
	return matchers.Must2(a.CreateToken())
}

func (a API) TestToken(token string) (*http.Response, error) {
	return req.Get(
		a.baseURL+"/tokentest",
		map[string]string{"Authorization": token},
	)
}

func (a API) MustTestToken(token string) *http.Response {
	return matchers.Must2(a.TestToken(token))
}
