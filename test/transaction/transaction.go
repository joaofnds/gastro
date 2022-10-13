package transaction

import (
	"astro/habit"
	"context"
	"database/sql"

	"go.uber.org/fx"
)

var Module = fx.Module("transaction", fx.Invoke(HookTransaction))

func HookTransaction(lc fx.Lifecycle, db *sql.DB, repo *habit.SQLRepository) {
	var transaction *sql.Tx

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			var err error
			transaction, err = db.BeginTx(context.Background(), nil)
			if err != nil {
				return err
			}

			repo.DB = transaction
			return nil
		},
		OnStop: func(context.Context) error {
			return transaction.Rollback()
		},
	})
}
