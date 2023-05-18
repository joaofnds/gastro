package postgres

import (
	"context"
	"database/sql"
	_ "embed"

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

//go:embed schema.sql
var schema string

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

			if _, err := db.ExecContext(ctx, schema); err != nil {
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
