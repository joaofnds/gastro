package habit

import "time"

type Habit struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	Activities []Activity `json:"activities"`
}

type Activity struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}
