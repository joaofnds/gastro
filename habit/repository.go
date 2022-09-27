package habit

import (
	"context"
	"time"
)

type HabitRepository interface {
	Create(ctx context.Context, userID, name string) (Habit, error)
	FindByName(ctx context.Context, userID, name string) (Habit, error)
	AddActivity(ctx context.Context, habit Habit, time time.Time) (Activity, error)
	List(ctx context.Context, userID string) ([]Habit, error)
	DeleteByName(ctx context.Context, userID string, name string) error
	DeleteAll(ctx context.Context) error
}
