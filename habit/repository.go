package habit

import (
	"context"
)

type HabitRepository interface {
	Create(ctx context.Context, create CreateHabitDTO) (Habit, error)
	List(ctx context.Context, userID string) ([]Habit, error)
	Find(ctx context.Context, find FindHabitDTO) (Habit, error)
	Update(ctx context.Context, dto UpdateHabitDTO) error
	Delete(ctx context.Context, find FindHabitDTO) error
	DeleteAll(ctx context.Context) error
}

type ActivityRepository interface {
	Add(ctx context.Context, habit Habit, dto AddActivityDTO) (Activity, error)
	Update(ctx context.Context, dto UpdateActivityDTO) (Activity, error)
	Find(ctx context.Context, find FindActivityDTO) (Activity, error)
	Delete(ctx context.Context, activity Activity) error
}

type GroupRepository interface {
	Create(ctx context.Context, dto CreateGroupDTO) (Group, error)
	Find(ctx context.Context, dto FindGroupDTO) (Group, error)
	Delete(ctx context.Context, group Group) error
	Join(ctx context.Context, habit Habit, group Group) error
	Leave(ctx context.Context, habit Habit, group Group) error
	GroupsAndHabits(ctx context.Context, userID string) ([]Group, []Habit, error)
}
