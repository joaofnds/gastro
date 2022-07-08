package postgres

import (
	"astro/config"
	"context"
	"database/sql"

	"go.uber.org/fx"
	"go.uber.org/zap"

	_ "github.com/lib/pq"
)

var Module = fx.Options(fx.Provide(NewClient), fx.Invoke(HookConnection))

func NewClient(config config.AppConfig, logger *zap.Logger) (*sql.DB, error) {
	db, err := sql.Open("postgres", config.Postgres.ConnectionString())
	if err != nil {
		logger.Error("failed to connect to postgres", zap.Error(err))
		return nil, err
	}

	return db, nil
}

func HookConnection(lifecycle fx.Lifecycle, db *sql.DB, logger *zap.Logger) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			err := db.PingContext(ctx)
			if err != nil {
				logger.Error("failed to ping db", zap.Error(err))
				return err
			}

			logger.Info("successfully pinged db")

			_, err = db.ExecContext(ctx, `
				CREATE TABLE IF NOT EXISTS habits (
					id   SERIAL PRIMARY KEY,
					name VARCHAR NOT NULL
				)
			`)

			if err != nil {
				logger.Error("failed to create astro table", zap.Error(err))
				return err
			}

			return nil
		},

		OnStop: func(ctx context.Context) error {
			err := db.Close()
			if err != nil {
				logger.Error("failed to close db connection", zap.Error(err))
				return err
			}

			return nil
		},
	})
}
