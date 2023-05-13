package habit

import "errors"

var (
	ErrNotFound   = errors.New("habit not found")
	ErrRepository = errors.New("repository error")
)
