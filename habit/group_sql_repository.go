package habit

import (
	"context"

	"gorm.io/gorm"
)

type GroupSQLRepository struct {
	orm *gorm.DB
}

func NewGroupSQLRepository(orm *gorm.DB) *GroupSQLRepository {
	return &GroupSQLRepository{orm}
}

func (repo *GroupSQLRepository) Create(ctx context.Context, dto CreateGroupDTO) (Group, error) {
	group := Group{Name: dto.Name, UserID: dto.UserID, Habits: []Habit{}}
	return group, resultErr(repo.orm.WithContext(ctx).Create(&group))
}

func (repo *GroupSQLRepository) Find(ctx context.Context, dto FindGroupDTO) (Group, error) {
	var group Group
	result := repo.orm.
		WithContext(ctx).
		Preload("Habits.Activities").
		First(&group, "id = ? and user_id = ?", dto.GroupID, dto.UserID)

	return group, resultErr(result)
}

func (repo *GroupSQLRepository) Delete(ctx context.Context, group Group) error {
	return resultErr(repo.orm.WithContext(ctx).Delete(&group))
}

func (repo *GroupSQLRepository) Join(ctx context.Context, habit Habit, group Group) error {
	return translateError(repo.orm.WithContext(ctx).Model(&group).Association("Habits").Append(&habit))
}

func (repo *GroupSQLRepository) Leave(ctx context.Context, habit Habit, group Group) error {
	return translateError(repo.orm.WithContext(ctx).Model(&group).Association("Habits").Delete(habit))
}

func (repo *GroupSQLRepository) GroupsAndHabits(ctx context.Context, userID string) ([]Group, []Habit, error) {
	var groups []Group
	var habits []Habit

	result := repo.orm.WithContext(ctx).Preload("Habits.Activities").Find(&groups, "user_id = ?", userID)
	if result.Error != nil {
		return groups, habits, translateError(result.Error)
	}

	result = repo.orm.
		Preload("Activities").
		Joins("LEFT JOIN groups_habits ON groups_habits.habit_id = habits.id").
		Where("groups_habits IS NULL").
		Find(&habits, "habits.user_id = ?", userID)

	return groups, habits, translateError(result.Error)
}
