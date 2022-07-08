package habit

import "context"

type HabitRepository interface {
	Create(ctx context.Context, name string) (Habit, error)
	FindByName(ctx context.Context, name string) (Habit, error)
	List(ctx context.Context) ([]Habit, error)
	DeleteByName(ctx context.Context, name string) error
	DeleteAll(ctx context.Context) error
}
