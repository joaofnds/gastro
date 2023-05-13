package habit

import "context"

type Habit struct {
	ID         string     `json:"id" gorm:"default:uuid_generate_v4()"`
	UserID     string     `json:"user_id"`
	Name       string     `json:"name"`
	Activities []Activity `json:"activities" gorm:"foreignKey:HabitID"`
}

type FindHabitDTO struct {
	HabitID string
	UserID  string
}

type CreateHabitDTO struct {
	Name   string
	UserID string
}

type UpdateHabitDTO struct {
	Name    string
	HabitID string
}

type FindActivityDTO struct {
	HabitID    string
	ActivityID string
	UserID     string
}

type HabitRepository interface {
	Create(ctx context.Context, create CreateHabitDTO) (Habit, error)
	List(ctx context.Context, userID string) ([]Habit, error)
	Find(ctx context.Context, find FindHabitDTO) (Habit, error)
	Update(ctx context.Context, dto UpdateHabitDTO) error
	Delete(ctx context.Context, find FindHabitDTO) error
	DeleteAll(ctx context.Context) error
}
