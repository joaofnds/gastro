package habit

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, create CreateDTO) (Habit, error)
	List(ctx context.Context, userID string) ([]Habit, error)
	Find(ctx context.Context, find FindDTO) (Habit, error)
	Delete(ctx context.Context, find FindDTO) error
	DeleteAll(ctx context.Context) error
	AddActivity(ctx context.Context, habit Habit, time time.Time) (Activity, error)
}
