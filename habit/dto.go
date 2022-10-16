package habit

type FindDTO struct {
	HabitID string
	UserID  string
}

type CreateDTO struct {
	Name   string
	UserID string
}

type FindActivityDTO struct {
	HabitID string
	ActivityID string
	UserID  string
}