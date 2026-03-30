package postgres

import (
	"context"
	"database/sql"
	"time"

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
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	return sqlDB, nil
}

func HookConnection(lifecycle fx.Lifecycle, db *sql.DB) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error { return db.PingContext(ctx) },
		OnStop:  func(ctx context.Context) error { return db.Close() },
	})
}
