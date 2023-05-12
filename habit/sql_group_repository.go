package habit

import (
	"context"

	"gorm.io/gorm"
)

type SQLGroupRepository struct {
	ORM *gorm.DB
}

func NewSQLGroupRepository(orm *gorm.DB) *SQLGroupRepository {
	return &SQLGroupRepository{orm}
}

func (repo *SQLGroupRepository) Create(ctx context.Context, dto CreateGroupDTO) (Group, error) {
	group := Group{Name: dto.Name, UserID: dto.UserID, Habits: []Habit{}}
	return group, resultErr(repo.ORM.Create(&group))
}

func (repo *SQLGroupRepository) Find(ctx context.Context, dto FindGroupDTO) (Group, error) {
	var group Group
	result := repo.ORM.
		WithContext(ctx).
		Preload("Habits.Activities").
		First(&group, "id = ? and user_id = ?", dto.GroupID, dto.UserID)

	return group, resultErr(result)
}

func (repo *SQLGroupRepository) Delete(ctx context.Context, group Group) error {
	return resultErr(repo.ORM.Delete(&group))
}

func (repo *SQLGroupRepository) Join(ctx context.Context, habit Habit, group Group) error {
	return translateError(repo.ORM.Model(&group).Association("Habits").Append(&habit))
}

func (repo *SQLGroupRepository) Leave(ctx context.Context, habit Habit, group Group) error {
	return translateError(repo.ORM.Model(&group).Association("Habits").Delete(habit))
}

func (repo *SQLGroupRepository) GroupsAndHabits(ctx context.Context, userID string) ([]Group, []Habit, error) {
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
