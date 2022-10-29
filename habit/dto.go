package habit

import "time"

type FindDTO struct {
	HabitID string
	UserID  string
}

type CreateDTO struct {
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
