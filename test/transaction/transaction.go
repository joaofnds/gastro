package transaction

import (
	"astro/habit"
	"context"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

var Module = fx.Module("transaction", fx.Invoke(HookTransaction))

func HookTransaction(
	lc fx.Lifecycle,
	db *gorm.DB,
	habitRepo *habit.SQLHabitRepository,
	activityRepo *habit.SQLActivityRepository,
	groupRepo *habit.SQLGroupRepository,
) {
	var tx *gorm.DB
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			tx = db.Begin()
			habitRepo.ORM = tx
			activityRepo.ORM = tx
			groupRepo.ORM = tx
			return nil
		},
		OnStop: func(context.Context) error {
			return tx.Rollback().Error
		},
	})
}
