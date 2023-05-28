package driver

import (
	"astro/habit"
	"io"
	"net/http"
	"time"
)

type Driver struct {
	api   *API
	Token string
}

func NewDriver(url string) *Driver {
	return &Driver{NewAPI(url), ""}
}

func (d *Driver) GetToken() {
	token, err := d.CreateToken()
	if err != nil {
		panic(err)
	}
	d.Token = token
}

func (d *Driver) List() ([]habit.Habit, error) {
	var data []habit.Habit

	return data, makeJSONRequest(params{
		into:   &data,
		status: http.StatusOK,
		req:    func() (*http.Response, error) { return d.api.List(d.Token) },
	})
}

func (d *Driver) Create(name string) (habit.Habit, error) {
	var habit habit.Habit

	return habit, makeJSONRequest(params{
		into:   &habit,
		status: http.StatusCreated,
		req:    func() (*http.Response, error) { return d.api.Create(d.Token, name) },
	})
}

func (d *Driver) Update(habitID, name string) error {
	return makeJSONRequest(params{
		status: http.StatusOK,
		req:    func() (*http.Response, error) { return d.api.Update(d.Token, habitID, name) },
	})
}

func (d *Driver) Get(id string) (habit.Habit, error) {
	var habit habit.Habit

	return habit, makeJSONRequest(params{
		into:   &habit,
		status: http.StatusOK,
		req:    func() (*http.Response, error) { return d.api.Get(d.Token, id) },
	})
}

func (d *Driver) AddActivity(id string, date time.Time) error {
	return makeJSONRequest(params{
		status: http.StatusCreated,
		req:    func() (*http.Response, error) { return d.api.AddActivity(d.Token, id, "", date) },
	})
}

func (d *Driver) AddActivityWithDesc(id, desc string, date time.Time) error {
	return makeJSONRequest(params{
		status: http.StatusCreated,
		req:    func() (*http.Response, error) { return d.api.AddActivity(d.Token, id, desc, date) },
	})
}

func (d *Driver) UpdateActivityDesc(habitID, activityID, desc string) error {
	return makeJSONRequest(params{
		status: http.StatusOK,
		req:    func() (*http.Response, error) { return d.api.UpdateActivity(d.Token, habitID, activityID, desc) },
	})
}

func (d *Driver) DeleteActivity(habitID, activityID string) error {
	return makeJSONRequest(params{
		status: http.StatusOK,
		req:    func() (*http.Response, error) { return d.api.DeleteActivity(d.Token, habitID, activityID) },
	})
}

func (d *Driver) CreateGroup(name string) (habit.Group, error) {
	var data habit.Group

	return data, makeJSONRequest(params{
		into:   &data,
		status: http.StatusCreated,
		req:    func() (*http.Response, error) { return d.api.CreateGroup(d.Token, name) },
	})
}

func (d *Driver) AddToGroup(habit habit.Habit, group habit.Group) error {
	return makeJSONRequest(params{
		status: http.StatusCreated,
		req:    func() (*http.Response, error) { return d.api.AddToGroup(d.Token, habit.ID, group.ID) },
	})
}

func (d *Driver) RemoveFromGroup(habit habit.Habit, group habit.Group) error {
	return makeJSONRequest(params{
		status: http.StatusOK,
		req:    func() (*http.Response, error) { return d.api.RemoveFromGroup(d.Token, habit.ID, group.ID) },
	})
}

func (d *Driver) DeleteGroup(group habit.Group) error {
	return makeJSONRequest(params{
		status: http.StatusOK,
		req:    func() (*http.Response, error) { return d.api.DeleteGroup(d.Token, group.ID) },
	})
}

func (d *Driver) GroupsAndHabits() ([]habit.Group, []habit.Habit, error) {
	var data habit.GroupsAndHabitsPayload

	return data.Groups, data.Habits, makeJSONRequest(params{
		into:   &data,
		status: http.StatusOK,
		req:    func() (*http.Response, error) { return d.api.GroupsAndHabits(d.Token) },
	})
}

func (d *Driver) CreateToken() (string, error) {
	res, err := d.api.CreateToken()
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	return string(b), err
}
