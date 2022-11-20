package habit

import "time"

type Group struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Habits []Habit `json:"habits"`
}

type Habit struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	Name       string     `json:"name"`
	Activities []Activity `json:"activities"`
}

type Activity struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Desc      string    `json:"description"`
}
