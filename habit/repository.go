package habit

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, create CreateDTO) (Habit, error)
	List(ctx context.Context, userID string) ([]Habit, error)
	Find(ctx context.Context, find FindDTO) (Habit, error)
	Delete(ctx context.Context, find FindDTO) error
	DeleteAll(ctx context.Context) error
	AddActivity(ctx context.Context, habit Habit, dto AddActivityDTO) (Activity, error)
	FindActivity(ctx context.Context, find FindActivityDTO) (Activity, error)
	DeleteActivity(ctx context.Context, activity Activity) error
}
