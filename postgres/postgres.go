package postgres

import (
	"context"
	"database/sql"

	"go.uber.org/fx"
	"go.uber.org/zap"

	_ "github.com/lib/pq"
)

var Module = fx.Module("postgres", fx.Provide(NewClient), fx.Invoke(HookConnection))

func NewClient(postgresConfig Config, logger *zap.Logger) (*sql.DB, error) {
	db, err := sql.Open("postgres", postgresConfig.ConnectionString())
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

			if err = enableUUIDExtension(ctx, db); err != nil {
				logger.Error("failed to enable uuid extension", zap.Error(err))
				return err
			}

			if err = createHabitsTable(ctx, db); err != nil {
				logger.Error("failed to create habits table", zap.Error(err))
				return err
			}

			if err = createActivitiesTable(ctx, db); err != nil {
				logger.Error("failed to create activities table", zap.Error(err))
				return err
			}

			if err = createGroupsTable(ctx, db); err != nil {
				logger.Error("failed to create groups table", zap.Error(err))
				return err
			}

			if err = createGroupsHabitsJoinTable(ctx, db); err != nil {
				logger.Error("failed to create groups-habits join table", zap.Error(err))
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
			id   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID NOT NULL,
			name VARCHAR NOT NULL
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_habit_id_and_user_id ON habits(id, user_id);
	`)
	return err
}

func createActivitiesTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS activities (
			id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			habit_id    UUID NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
			description VARCHAR NOT NULL DEFAULT '',
			created_at  TIMESTAMP NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_activity_habit ON activities(habit_id);
	`)
	return err
}

func createGroupsTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS "public"."groups" (
            id      UUID NOT NULL DEFAULT uuid_generate_v4(),
            name    VARCHAR NOT NULL,
			user_id UUID NOT NULL,
            PRIMARY KEY (id, user_id)
        );
    `)
	return err
}

func createGroupsHabitsJoinTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS groups_habits (
            group_id UUID NOT NULL,
            habit_id UUID NOT NULL,
            user_id  UUID NOT NULL,
            CONSTRAINT "groups_habits_habit_id_fkey" FOREIGN KEY (habit_id, user_id) REFERENCES habits (id, user_id) ON DELETE RESTRICT ON UPDATE CASCADE,
            CONSTRAINT "groups_habits_group_id_fkey" FOREIGN KEY (group_id, user_id) REFERENCES groups (id, user_id) ON DELETE RESTRICT ON UPDATE CASCADE,
            PRIMARY KEY (group_id, habit_id)
        );
    `)
	return err
}

func enableUUIDExtension(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	return err
}
