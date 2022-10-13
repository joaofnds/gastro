package habit

import "errors"

var (
	NotFoundErr   = errors.New("habit not found")
	RepositoryErr = errors.New("repository error")
)
