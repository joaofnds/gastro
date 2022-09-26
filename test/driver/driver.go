package driver

import (
	"astro/habit"
	"encoding/json"
	"io"
)

type Driver struct {
	api *API
}

func NewDriver() *Driver {
	return &Driver{NewAPI()}
}

func (d *Driver) List() ([]habit.Habit, error) {
	data := []habit.Habit{}

	res, err := d.api.List()
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
	res, err := d.api.Create(name)
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

func (d *Driver) Get(name string) (habit.Habit, error) {
	data := habit.Habit{}

	res, err := d.api.Get(name)
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

func (d *Driver) CreateToken() ([]byte, error) {
	res, err := d.api.CreateToken()
	if err != nil {
		return []byte{}, err
	}
	defer res.Body.Close()

	return io.ReadAll(res.Body)
}
