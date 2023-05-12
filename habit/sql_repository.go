package habit

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type SQLRepository struct {
	ORM *gorm.DB
}

func NewSQLRepository(orm *gorm.DB) *SQLRepository {
	return &SQLRepository{orm}
}

func (repo *SQLRepository) Create(ctx context.Context, create CreateHabitDTO) (Habit, error) {
	var habit Habit
	habit.UserID = create.UserID
	habit.Name = create.Name
	habit.Activities = []Activity{}
	return habit, resultErr(repo.ORM.WithContext(ctx).Create(&habit))
}

func (repo *SQLRepository) Find(ctx context.Context, find FindHabitDTO) (Habit, error) {
	var habit Habit
	result := repo.ORM.
		WithContext(ctx).
		Preload("Activities").
		First(&habit, "id = ? and user_id = ?", find.HabitID, find.UserID)

	return habit, resultErr(result)
}

func (repo *SQLRepository) Update(ctx context.Context, dto UpdateHabitDTO) error {
	return resultErr(repo.ORM.WithContext(ctx).Select("Name").Updates(&Habit{ID: dto.HabitID, Name: dto.Name}))
}

func (repo *SQLRepository) AddActivity(ctx context.Context, habit Habit, dto AddActivityDTO) (Activity, error) {
	activity := Activity{Desc: dto.Desc, CreatedAt: dto.Time}
	return activity, translateError(repo.ORM.WithContext(ctx).Model(&habit).Association("Activities").Append(&activity))
}

func (repo *SQLRepository) FindActivity(ctx context.Context, find FindActivityDTO) (Activity, error) {
	var activity Activity
	result := repo.ORM.WithContext(ctx).First(&activity, "habit_id = ? and id = ?", find.HabitID, find.ActivityID)
	return activity, translateError(result.Error)
}

func (repo *SQLRepository) UpdateActivity(ctx context.Context, dto UpdateActivityDTO) (Activity, error) {
	activity := Activity{ID: dto.ActivityID, Desc: dto.Desc}
	return activity, resultErr(repo.ORM.WithContext(ctx).Select("Desc").Updates(&activity))
}

func (repo *SQLRepository) DeleteActivity(ctx context.Context, activity Activity) error {
	return resultErr(repo.ORM.Delete(&activity))
}

func (repo *SQLRepository) List(ctx context.Context, userID string) ([]Habit, error) {
	var habits []Habit
	return habits, translateError(repo.ORM.WithContext(ctx).Preload("Activities").Find(&habits).Error)
}

func (repo *SQLRepository) Delete(ctx context.Context, find FindHabitDTO) error {
	return resultErr(repo.ORM.WithContext(ctx).Where("user_id = ?", find.UserID).Delete(&Habit{ID: find.HabitID}))
}

func (repo *SQLRepository) DeleteAll(ctx context.Context) error {
	return resultErr(repo.ORM.WithContext(ctx).Exec("DELETE FROM habits"))
}

func (repo *SQLRepository) CreateGroup(ctx context.Context, dto CreateGroupDTO) (Group, error) {
	group := Group{Name: dto.Name, UserID: dto.UserID, Habits: []Habit{}}
	return group, resultErr(repo.ORM.Create(&group))
}

func (repo *SQLRepository) AddToGroup(ctx context.Context, habit Habit, group Group) error {
	return translateError(repo.ORM.Model(&group).Association("Habits").Append(&habit))
}

func (repo *SQLRepository) RemoveFromGroup(ctx context.Context, habit Habit, group Group) error {
	return translateError(repo.ORM.Model(&group).Association("Habits").Delete(habit))
}

func (repo *SQLRepository) DeleteGroup(ctx context.Context, group Group) error {
	return resultErr(repo.ORM.Delete(&group))
}

func (repo *SQLRepository) GroupsAndHabits(ctx context.Context, userID string) ([]Group, []Habit, error) {
	var groups []Group
	var habits []Habit

	result := repo.ORM.WithContext(ctx).Preload("Habits.Activities").Find(&groups, "user_id = ?", userID)
	if result.Error != nil {
		return groups, habits, translateError(result.Error)
	}

	result = repo.ORM.
		Preload("Activities").
		Joins("LEFT JOIN groups_habits ON groups_habits.habit_id = habits.id").
		Where("groups_habits IS NULL").
		Find(&habits, "habits.user_id = ?", userID)

	return groups, habits, translateError(result.Error)
}

func (repo *SQLRepository) FindGroup(ctx context.Context, dto FindGroupDTO) (Group, error) {
	var group Group
	result := repo.ORM.
		WithContext(ctx).
		Preload("Habits.Activities").
		First(&group, "id = ? and user_id = ?", dto.GroupID, dto.UserID)

	return group, resultErr(result)
}

func resultErr(result *gorm.DB) error {
	err := result.Error
	if err == nil && result.RowsAffected == 0 {
		return ErrNotFound
	}
	return translateError(err)
}

func translateError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return ErrNotFound
	default:
		fmt.Printf("\n%v\n", err)
		return ErrRepository
	}
}
