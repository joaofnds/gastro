package habit

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, create CreateHabitDTO) (Habit, error)
	List(ctx context.Context, userID string) ([]Habit, error)
	Find(ctx context.Context, find FindHabitDTO) (Habit, error)
	Update(ctx context.Context, dto UpdateHabitDTO) error
	Delete(ctx context.Context, find FindHabitDTO) error
	DeleteAll(ctx context.Context) error
	AddActivity(ctx context.Context, habit Habit, dto AddActivityDTO) (Activity, error)
	UpdateActivity(ctx context.Context, dto UpdateActivityDTO) (Activity, error)
	FindActivity(ctx context.Context, find FindActivityDTO) (Activity, error)
	DeleteActivity(ctx context.Context, activity Activity) error
}
