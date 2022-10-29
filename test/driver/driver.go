package driver

import (
	"astro/habit"
	"encoding/json"
	"fmt"
	"io"
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
	data := []habit.Habit{}

	res, err := d.api.List(d.Token)
	if err != nil {
		return data, err
	}
	defer res.Body.Close()

	str, err := io.ReadAll(res.Body)
	if err != nil {
		return data, err
	}

	err = json.Unmarshal(str, &data)

	return data, err
}

func (d *Driver) Create(name string) (habit.Habit, error) {
	var habit habit.Habit
	res, err := d.api.Create(d.Token, name)
	if err != nil {
		return habit, err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return habit, err
	}
	defer res.Body.Close()

	err = json.Unmarshal(body, &habit)
	return habit, err
}

func (d *Driver) Get(id string) (habit.Habit, error) {
	data := habit.Habit{}

	res, err := d.api.Get(d.Token, id)
	if err != nil {
		return data, err
	}
	defer res.Body.Close()

	str, err := io.ReadAll(res.Body)
	if err != nil {
		return data, err
	}

	err = json.Unmarshal(str, &data)

	return data, err
}

func (d *Driver) AddActivity(id string) error {
	res, err := d.api.AddActivity(d.Token, id, "")
	if err != nil {
		return err
	}

	if res.StatusCode != 201 {
		return fmt.Errorf("failed to add activity")
	}

	return nil
}

func (d *Driver) AddActivityWithDesc(id, desc string) error {
	res, err := d.api.AddActivity(d.Token, id, desc)
	if err != nil {
		return err
	}

	if res.StatusCode != 201 {
		return fmt.Errorf("failed to add activity")
	}

	return nil
}

func (d *Driver) UpdateActivityDesc(habitID, activityID, desc string) error {
	res, err := d.api.UpdateActivity(d.Token, habitID, activityID, desc)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("failed to update activity")
	}

	return nil
}

func (d *Driver) DeleteActivity(habitID, activityID string) error {
	res, err := d.api.DeleteActivity(d.Token, habitID, activityID)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("failed to delete activity")
	}

	return nil
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
