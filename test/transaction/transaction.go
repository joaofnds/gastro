package transaction

import (
	"astro/habit"
	"context"
	"database/sql"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

var Module = fx.Module("transaction", fx.Invoke(HookTransaction))

func HookTransaction(lc fx.Lifecycle, db *sql.DB, orm *gorm.DB, repo *habit.SQLRepository) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			repo.ORM = orm.Begin()
			return nil
		},
		OnStop: func(context.Context) error {
			return repo.ORM.Rollback().Error
		},
	})
}
