package postgres

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Module = fx.Module(
	"postgres",
	fx.Provide(NewGormDB),
	fx.Provide(NewSQLDB),
	fx.Provide(NewHealthChecker),
	fx.Invoke(HookConnection),
)

func NewGormDB(postgresConfig Config) (*gorm.DB, error) {
	return gorm.Open(
		postgres.Open(postgresConfig.Addr),
		&gorm.Config{Logger: nil},
	)
}

func NewSQLDB(db *gorm.DB) (*sql.DB, error) {
	return db.DB()
}

func HookConnection(lifecycle fx.Lifecycle, db *sql.DB) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error { return db.PingContext(ctx) },
		OnStop:  func(ctx context.Context) error { return db.Close() },
	})
}
