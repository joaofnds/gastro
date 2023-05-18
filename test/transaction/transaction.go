package transaction

import (
	"context"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

var Module = fx.Decorate(HookTransaction)

func HookTransaction(lc fx.Lifecycle, db *gorm.DB) *gorm.DB {
	tx := db.Begin()
	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			return tx.Rollback().Error
		},
	})
	return tx
}
