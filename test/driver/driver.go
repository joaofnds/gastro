package driver

import (
	"astro/habit"
	"astro/test/matchers"
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

func (d *Driver) MustList() []habit.Habit {
	return matchers.Must2(d.List())
}

func (d *Driver) Create(name string) (habit.Habit, error) {
	var habit habit.Habit

	return habit, makeJSONRequest(params{
		into:   &habit,
		status: http.StatusCreated,
		req:    func() (*http.Response, error) { return d.api.Create(d.Token, name) },
	})
}

func (d *Driver) MustCreate(name string) habit.Habit {
	return matchers.Must2(d.Create(name))
}

func (d *Driver) Update(habitID, name string) error {
	return makeJSONRequest(params{
		status: http.StatusOK,
		req:    func() (*http.Response, error) { return d.api.Update(d.Token, habitID, name) },
	})
}

func (d *Driver) MustUpdate(habitID, name string) {
	matchers.Must(d.Update(habitID, name))
}

func (d *Driver) Get(id string) (habit.Habit, error) {
	var habit habit.Habit

	return habit, makeJSONRequest(params{
		into:   &habit,
		status: http.StatusOK,
		req:    func() (*http.Response, error) { return d.api.Get(d.Token, id) },
	})
}

func (d *Driver) MustGet(id string) habit.Habit {
	return matchers.Must2(d.Get(id))
}

func (d *Driver) AddActivity(id string, date time.Time) error {
	return makeJSONRequest(params{
		status: http.StatusCreated,
		req:    func() (*http.Response, error) { return d.api.AddActivity(d.Token, id, "", date) },
	})
}

func (d *Driver) MustAddActivity(id string, date time.Time) {
	matchers.Must(d.AddActivity(id, date))
}

func (d *Driver) AddActivityWithDesc(id, desc string, date time.Time) error {
	return makeJSONRequest(params{
		status: http.StatusCreated,
		req:    func() (*http.Response, error) { return d.api.AddActivity(d.Token, id, desc, date) },
	})
}

func (d *Driver) MustAddActivityWithDesc(id, desc string, date time.Time) {
	matchers.Must(d.AddActivityWithDesc(id, desc, date))
}

func (d *Driver) UpdateActivityDesc(habitID, activityID, desc string) error {
	return makeJSONRequest(params{
		status: http.StatusOK,
		req:    func() (*http.Response, error) { return d.api.UpdateActivity(d.Token, habitID, activityID, desc) },
	})
}

func (d *Driver) MustUpdateActivityDesc(habitID, activityID, desc string) {
	matchers.Must(d.UpdateActivityDesc(habitID, activityID, desc))
}

func (d *Driver) DeleteActivity(habitID, activityID string) error {
	return makeJSONRequest(params{
		status: http.StatusOK,
		req:    func() (*http.Response, error) { return d.api.DeleteActivity(d.Token, habitID, activityID) },
	})
}

func (d *Driver) MustDeleteActivity(habitID, activityID string) {
	matchers.Must(d.DeleteActivity(habitID, activityID))
}

func (d *Driver) CreateGroup(name string) (habit.Group, error) {
	var data habit.Group

	return data, makeJSONRequest(params{
		into:   &data,
		status: http.StatusCreated,
		req:    func() (*http.Response, error) { return d.api.CreateGroup(d.Token, name) },
	})
}

func (d *Driver) MustCreateGroup(name string) habit.Group {
	return matchers.Must2(d.CreateGroup(name))
}

func (d *Driver) AddToGroup(habit habit.Habit, group habit.Group) error {
	return makeJSONRequest(params{
		status: http.StatusCreated,
		req:    func() (*http.Response, error) { return d.api.AddToGroup(d.Token, habit.ID, group.ID) },
	})
}

func (d *Driver) MustAddToGroup(habit habit.Habit, group habit.Group) {
	matchers.Must(d.AddToGroup(habit, group))
}

func (d *Driver) RemoveFromGroup(habit habit.Habit, group habit.Group) error {
	return makeJSONRequest(params{
		status: http.StatusOK,
		req:    func() (*http.Response, error) { return d.api.RemoveFromGroup(d.Token, habit.ID, group.ID) },
	})
}

func (d *Driver) MustRemoveFromGroup(habit habit.Habit, group habit.Group) {
	matchers.Must(d.RemoveFromGroup(habit, group))
}

func (d *Driver) DeleteGroup(group habit.Group) error {
	return makeJSONRequest(params{
		status: http.StatusOK,
		req:    func() (*http.Response, error) { return d.api.DeleteGroup(d.Token, group.ID) },
	})
}

func (d *Driver) MustDeleteGroup(group habit.Group) {
	matchers.Must(d.DeleteGroup(group))
}

func (d *Driver) GroupsAndHabits() ([]habit.Group, []habit.Habit, error) {
	var data habit.GroupsAndHabitsPayload

	return data.Groups, data.Habits, makeJSONRequest(params{
		into:   &data,
		status: http.StatusOK,
		req:    func() (*http.Response, error) { return d.api.GroupsAndHabits(d.Token) },
	})
}

func (d *Driver) MustGroupsAndHabits() ([]habit.Group, []habit.Habit) {
	return matchers.Must3(d.GroupsAndHabits())
}

func (d *Driver) CreateToken() (string, error) {
	res, err := d.api.CreateToken()
	if err != nil {
		return "", err
	}
	defer func() { _ = res.Body.Close() }()

	b, err := io.ReadAll(res.Body)
	return string(b), err
}

func (d *Driver) MustCreateToken() string {
	return matchers.Must2(d.CreateToken())
}
