package habit

import (
	"context"

	"gorm.io/gorm"
)

type HabitSQLRepository struct {
	ORM *gorm.DB
}

func NewHabitSQLRepository(orm *gorm.DB) *HabitSQLRepository {
	return &HabitSQLRepository{orm}
}

func (repo *HabitSQLRepository) Create(ctx context.Context, create CreateHabitDTO) (Habit, error) {
	var habit Habit
	habit.UserID = create.UserID
	habit.Name = create.Name
	habit.Activities = []Activity{}
	return habit, resultErr(repo.ORM.WithContext(ctx).Create(&habit))
}

func (repo *HabitSQLRepository) List(ctx context.Context, userID string) ([]Habit, error) {
	var habits []Habit
	result := repo.ORM.
		WithContext(ctx).
		Preload("Activities").
		Find(&habits, "user_id = ?", userID)
	return habits, translateError(result.Error)
}

func (repo *HabitSQLRepository) Find(ctx context.Context, find FindHabitDTO) (Habit, error) {
	var habit Habit
	result := repo.ORM.
		WithContext(ctx).
		Preload("Activities").
		First(&habit, "id = ? and user_id = ?", find.HabitID, find.UserID)

	return habit, translateError(result.Error)
}

func (repo *HabitSQLRepository) Update(ctx context.Context, dto UpdateHabitDTO) error {
	return resultErr(repo.ORM.WithContext(ctx).Select("Name").Updates(&Habit{ID: dto.HabitID, Name: dto.Name}))
}

func (repo *HabitSQLRepository) Delete(ctx context.Context, find FindHabitDTO) error {
	return resultErr(repo.ORM.WithContext(ctx).Where("user_id = ?", find.UserID).Delete(&Habit{ID: find.HabitID}))
}

func (repo *HabitSQLRepository) DeleteAll(ctx context.Context) error {
	return resultErr(repo.ORM.WithContext(ctx).Exec("DELETE FROM habits"))
}
