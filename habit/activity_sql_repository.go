package habit

import (
	"context"

	"gorm.io/gorm"
)

type SQLActivityRepository struct {
	ORM *gorm.DB
}

func NewSQLActivityRepository(orm *gorm.DB) *SQLActivityRepository {
	return &SQLActivityRepository{orm}
}

func (repo *SQLActivityRepository) Add(ctx context.Context, habit Habit, dto AddActivityDTO) (Activity, error) {
	activity := Activity{Desc: dto.Desc, CreatedAt: dto.Time}
	return activity, translateError(repo.ORM.WithContext(ctx).Model(&habit).Association("Activities").Append(&activity))
}

func (repo *SQLActivityRepository) Find(ctx context.Context, find FindActivityDTO) (Activity, error) {
	var activity Activity
	result := repo.ORM.WithContext(ctx).First(&activity, "habit_id = ? and id = ?", find.HabitID, find.ActivityID)
	return activity, translateError(result.Error)
}

func (repo *SQLActivityRepository) Update(ctx context.Context, dto UpdateActivityDTO) (Activity, error) {
	activity := Activity{ID: dto.ActivityID, Desc: dto.Desc}
	return activity, resultErr(repo.ORM.WithContext(ctx).Select("Desc").Updates(&activity))
}

func (repo *SQLActivityRepository) Delete(ctx context.Context, activity Activity) error {
	return resultErr(repo.ORM.WithContext(ctx).Delete(&activity))
}
