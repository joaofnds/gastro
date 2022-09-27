package postgres

import (
	"astro/config"
	"context"
	"database/sql"

	"go.uber.org/fx"
	"go.uber.org/zap"

	_ "github.com/lib/pq"
)

var Module = fx.Module("postgres", fx.Provide(NewClient), fx.Invoke(HookConnection))

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

			err = createHabitsTable(ctx, db)
			if err != nil {
				logger.Error("failed to create habits table", zap.Error(err))
				return err
			}

			err = createActivitiesTable(ctx, db)
			if err != nil {
				logger.Error("failed to create activities table", zap.Error(err))
				return err
			}

			err = enableUUIDExtension(ctx, db)
			if err != nil {
				logger.Error("failed to enable uuid extension", zap.Error(err))
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

func createHabitsTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS habits (
			id   SERIAL PRIMARY KEY,
			name VARCHAR NOT NULL,
			user_id UUID NOT NULL
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_habit_name ON habits(name, user_id);
	`)
	return err
}

func createActivitiesTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS activities (
			id         SERIAL PRIMARY KEY,
			habit_id   INTEGER NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_activity_habit ON activities(habit_id);
	`)
	return err
}

func enableUUIDExtension(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	return err
}
