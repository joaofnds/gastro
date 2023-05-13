package habit

import (
	"context"
	"time"
)

type Activity struct {
	ID        string    `json:"id" gorm:"default:uuid_generate_v4()"`
	CreatedAt time.Time `json:"created_at"`
	Desc      string    `json:"description" gorm:"column:description"`
	HabitID   string
}

type AddActivityDTO struct {
	Desc string
	Time time.Time
}

type UpdateActivityDTO struct {
	ActivityID string
	Desc       string
}

type ActivityRepository interface {
	Add(ctx context.Context, habit Habit, dto AddActivityDTO) (Activity, error)
	Update(ctx context.Context, dto UpdateActivityDTO) (Activity, error)
	Find(ctx context.Context, find FindActivityDTO) (Activity, error)
	Delete(ctx context.Context, activity Activity) error
}
