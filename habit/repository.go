package habit

import (
	"context"
	"time"
)

type HabitRepository interface {
	Create(ctx context.Context, name string) (Habit, error)
	FindByName(ctx context.Context, name string) (Habit, error)
	AddActivity(ctx context.Context, habit Habit, time time.Time) (Activity, error)
	List(ctx context.Context) ([]Habit, error)
	DeleteByName(ctx context.Context, name string) error
	DeleteAll(ctx context.Context) error
}
