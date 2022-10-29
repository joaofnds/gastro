package habit

import "time"

type FindHabitDTO struct {
	HabitID string
	UserID  string
}

type CreateHabitDTO struct {
	Name   string
	UserID string
}

type FindActivityDTO struct {
	HabitID    string
	ActivityID string
	UserID     string
}

type AddActivityDTO struct {
	Desc string
	Time time.Time
}

type UpdateActivityDTO struct {
	ActivityID string
	Desc       string
}
