package habit

import "errors"

var (
	HabitNotFoundErr = errors.New("habit not found")
	RepositoryErr    = errors.New("repository error")
)
