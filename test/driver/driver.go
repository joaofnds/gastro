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

func NewDriver() *Driver {
	return &Driver{NewAPI(), ""}
}

func (d *Driver) GetToken() {
	res, err := d.api.CreateToken()
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	token, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	d.Token = string(token)
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

func (d *Driver) Get(name string) (habit.Habit, error) {
	data := habit.Habit{}

	res, err := d.api.Get(d.Token, name)
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

func (d *Driver) AddActivity(name string) error {
	res, err := d.api.AddActivity(d.Token, name)
	if err != nil {
		return err
	}

	if res.StatusCode != 201 {
		return fmt.Errorf("failed to add activity")
	}

	return nil
}

func (d *Driver) CreateToken() ([]byte, error) {
	res, err := d.api.CreateToken()
	if err != nil {
		return []byte{}, err
	}
	defer res.Body.Close()

	return io.ReadAll(res.Body)
}
