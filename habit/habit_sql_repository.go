package habit

import (
	"context"

	"gorm.io/gorm"
)

type HabitSQLRepository struct {
	orm         *gorm.DB
	idGenerator IDGenerator
}

func NewHabitSQLRepository(orm *gorm.DB, idGenerator IDGenerator) *HabitSQLRepository {
	return &HabitSQLRepository{orm, idGenerator}
}

func (repo *HabitSQLRepository) Create(ctx context.Context, create CreateHabitDTO) (Habit, error) {
	if !normalizeUUIDs(&create.UserID) {
		return Habit{}, ErrRepository
	}

	id, err := repo.idGenerator.NewID()
	if err != nil {
		return Habit{}, ErrRepository
	}

	var habit Habit
	habit.ID = id
	habit.UserID = create.UserID
	habit.Name = create.Name
	habit.Activities = []Activity{}
	return habit, resultErr(repo.orm.WithContext(ctx).Create(&habit))
}

func (repo *HabitSQLRepository) List(ctx context.Context, userID string) ([]Habit, error) {
	if !normalizeUUIDs(&userID) {
		return nil, ErrRepository
	}

	var habits []Habit
	result := repo.orm.
		WithContext(ctx).
		Preload("Activities").
		Find(&habits, "user_id = ?", userID)
	return habits, translateError(result.Error)
}

func (repo *HabitSQLRepository) Find(ctx context.Context, find FindHabitDTO) (Habit, error) {
	if !normalizeUUIDs(&find.HabitID, &find.UserID) {
		return Habit{}, ErrRepository
	}

	var habit Habit
	result := repo.orm.
		WithContext(ctx).
		Preload("Activities").
		First(&habit, "id = ? and user_id = ?", find.HabitID, find.UserID)

	return habit, translateError(result.Error)
}

func (repo *HabitSQLRepository) Update(ctx context.Context, dto UpdateHabitDTO) error {
	if !normalizeUUIDs(&dto.HabitID) {
		return ErrRepository
	}

	return resultErr(repo.orm.WithContext(ctx).Select("Name").Updates(&Habit{ID: dto.HabitID, Name: dto.Name}))
}

func (repo *HabitSQLRepository) Delete(ctx context.Context, find FindHabitDTO) error {
	if !normalizeUUIDs(&find.HabitID, &find.UserID) {
		return ErrRepository
	}

	return resultErr(repo.orm.WithContext(ctx).Where("user_id = ?", find.UserID).Delete(&Habit{ID: find.HabitID}))
}

func (repo *HabitSQLRepository) DeleteAll(ctx context.Context) error {
	return resultErr(repo.orm.WithContext(ctx).Exec("DELETE FROM habits"))
}
