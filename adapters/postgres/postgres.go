package postgres

import (
	"context"
	"database/sql"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/lib/pq"
)

var Module = fx.Module(
	"postgres",
	fx.Provide(NewGormDB),
	fx.Provide(NewSQLDB),
	fx.Provide(NewHealthChecker),
	fx.Invoke(HookConnection),
)

func NewGormDB(postgresConfig Config, logger *zap.Logger) (*gorm.DB, error) {
	return gorm.Open(
		postgres.Open(postgresConfig.ConnectionString()),
		&gorm.Config{Logger: nil},
	)
}

func NewSQLDB(db *gorm.DB) (*sql.DB, error) {
	return db.DB()
}

func HookConnection(lifecycle fx.Lifecycle, db *sql.DB, logger *zap.Logger) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := db.PingContext(ctx); err != nil {
				logger.Error("failed to ping db", zap.Error(err))
				return err
			}
			logger.Info("successfully pinged db")

			if err := createTables(ctx, db); err != nil {
				logger.Error("failed to create habits table", zap.Error(err))
				return err
			}

			return nil
		},

		OnStop: func(ctx context.Context) error {
			if err := db.Close(); err != nil {
				logger.Error("failed to close db connection", zap.Error(err))
				return err
			}

			return nil
		},
	})
}

func createTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

    CREATE TABLE IF NOT EXISTS habits (
      id   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
      user_id UUID NOT NULL,
      name VARCHAR NOT NULL
    );
    CREATE UNIQUE INDEX IF NOT EXISTS idx_habit_id_and_user_id ON habits(id, user_id);

    CREATE TABLE IF NOT EXISTS activities (
      id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
      habit_id    UUID NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
      description VARCHAR NOT NULL DEFAULT '',
      created_at  TIMESTAMP NOT NULL DEFAULT NOW()
    );
    CREATE INDEX IF NOT EXISTS idx_activity_habit ON activities(habit_id);

    CREATE TABLE IF NOT EXISTS "public"."groups" (
      id      UUID NOT NULL DEFAULT uuid_generate_v4(),
      name    VARCHAR NOT NULL,
      user_id UUID NOT NULL,
      PRIMARY KEY (id, user_id)
    );

    CREATE TABLE IF NOT EXISTS groups_habits (
      group_id UUID NOT NULL,
      habit_id UUID NOT NULL,
      user_id  UUID NOT NULL,
      CONSTRAINT "groups_habits_habit_id_fkey" FOREIGN KEY (habit_id, user_id) REFERENCES habits (id, user_id) ON DELETE CASCADE ON UPDATE CASCADE,
      CONSTRAINT "groups_habits_group_id_fkey" FOREIGN KEY (group_id, user_id) REFERENCES groups (id, user_id) ON DELETE CASCADE ON UPDATE CASCADE,
      PRIMARY KEY (group_id, habit_id)
    );
  `)
	return err
}
