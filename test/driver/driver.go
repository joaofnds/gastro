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

func (d *Driver) Create(name string) error {
	_, err := d.api.Create(name)
	return err
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
