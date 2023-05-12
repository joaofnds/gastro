package habit

import (
	"context"

	"gorm.io/gorm"
)

type SQLHabitRepository struct {
	ORM *gorm.DB
}

func NewSQLHabitRepository(orm *gorm.DB) *SQLHabitRepository {
	return &SQLHabitRepository{orm}
}

func (repo *SQLHabitRepository) Create(ctx context.Context, create CreateHabitDTO) (Habit, error) {
	var habit Habit
	habit.UserID = create.UserID
	habit.Name = create.Name
	habit.Activities = []Activity{}
	return habit, resultErr(repo.ORM.WithContext(ctx).Create(&habit))
}

func (repo *SQLHabitRepository) List(ctx context.Context, userID string) ([]Habit, error) {
	var habits []Habit
	return habits, translateError(repo.ORM.WithContext(ctx).Preload("Activities").Find(&habits).Error)
}

func (repo *SQLHabitRepository) Find(ctx context.Context, find FindHabitDTO) (Habit, error) {
	var habit Habit
	result := repo.ORM.
		WithContext(ctx).
		Preload("Activities").
		First(&habit, "id = ? and user_id = ?", find.HabitID, find.UserID)

	return habit, resultErr(result)
}

func (repo *SQLHabitRepository) Update(ctx context.Context, dto UpdateHabitDTO) error {
	return resultErr(repo.ORM.WithContext(ctx).Select("Name").Updates(&Habit{ID: dto.HabitID, Name: dto.Name}))
}

func (repo *SQLHabitRepository) Delete(ctx context.Context, find FindHabitDTO) error {
	return resultErr(repo.ORM.WithContext(ctx).Where("user_id = ?", find.UserID).Delete(&Habit{ID: find.HabitID}))
}

func (repo *SQLHabitRepository) DeleteAll(ctx context.Context) error {
	return resultErr(repo.ORM.WithContext(ctx).Exec("DELETE FROM habits"))
}
