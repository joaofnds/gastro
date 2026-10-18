package habit

import (
	"context"

	"gorm.io/gorm"
)

type ActivitySQLRepository struct {
	orm         *gorm.DB
	idGenerator IDGenerator
}

func NewActivitySQLRepository(orm *gorm.DB, idGenerator IDGenerator) *ActivitySQLRepository {
	return &ActivitySQLRepository{orm, idGenerator}
}

func (repo *ActivitySQLRepository) Add(ctx context.Context, habit Habit, dto AddActivityDTO) (Activity, error) {
	if !normalizeUUIDs(&habit.ID) {
		return Activity{}, ErrRepository
	}

	id, err := repo.idGenerator.NewID()
	if err != nil {
		return Activity{}, ErrRepository
	}

	activity := Activity{ID: id, HabitID: habit.ID, Desc: dto.Desc, CreatedAt: dto.Time}
	return activity, translateError(repo.orm.WithContext(ctx).Create(&activity).Error)
}

func (repo *ActivitySQLRepository) Find(ctx context.Context, find FindActivityDTO) (Activity, error) {
	if !normalizeUUIDs(&find.HabitID, &find.ActivityID) {
		return Activity{}, ErrRepository
	}

	var activity Activity
	result := repo.orm.WithContext(ctx).First(&activity, "habit_id = ? and id = ?", find.HabitID, find.ActivityID)
	return activity, translateError(result.Error)
}

func (repo *ActivitySQLRepository) Update(ctx context.Context, dto UpdateActivityDTO) (Activity, error) {
	if !normalizeUUIDs(&dto.ActivityID) {
		return Activity{}, ErrRepository
	}

	activity := Activity{ID: dto.ActivityID, Desc: dto.Desc}
	return activity, resultErr(repo.orm.WithContext(ctx).Select("Desc").Updates(&activity))
}

func (repo *ActivitySQLRepository) Delete(ctx context.Context, activity Activity) error {
	if !normalizeUUIDs(&activity.ID) {
		return ErrRepository
	}

	return resultErr(repo.orm.WithContext(ctx).Delete(&activity))
}
