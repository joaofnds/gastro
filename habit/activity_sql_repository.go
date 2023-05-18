package habit

import (
	"context"

	"gorm.io/gorm"
)

type ActivitySQLRepository struct {
	orm *gorm.DB
}

func NewActivitySQLRepository(orm *gorm.DB) *ActivitySQLRepository {
	return &ActivitySQLRepository{orm}
}

func (repo *ActivitySQLRepository) Add(ctx context.Context, habit Habit, dto AddActivityDTO) (Activity, error) {
	activity := Activity{Desc: dto.Desc, CreatedAt: dto.Time}
	return activity, translateError(repo.orm.WithContext(ctx).Model(&habit).Association("Activities").Append(&activity))
}

func (repo *ActivitySQLRepository) Find(ctx context.Context, find FindActivityDTO) (Activity, error) {
	var activity Activity
	result := repo.orm.WithContext(ctx).First(&activity, "habit_id = ? and id = ?", find.HabitID, find.ActivityID)
	return activity, translateError(result.Error)
}

func (repo *ActivitySQLRepository) Update(ctx context.Context, dto UpdateActivityDTO) (Activity, error) {
	activity := Activity{ID: dto.ActivityID, Desc: dto.Desc}
	return activity, resultErr(repo.orm.WithContext(ctx).Select("Desc").Updates(&activity))
}

func (repo *ActivitySQLRepository) Delete(ctx context.Context, activity Activity) error {
	return resultErr(repo.orm.WithContext(ctx).Delete(&activity))
}
