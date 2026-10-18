package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/fx"
	driver "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite"
)

type Config struct {
	Path string `mapstructure:"path"`
}

var ErrUncommissioned = errors.New("database is not commissioned")

var Module = fx.Module("sqlite",
	fx.Provide(Open, NewSQLDB, NewHealthChecker),
	fx.Invoke(HookConnection),
)

func Open(config Config) (*gorm.DB, error) {
	info, err := os.Stat(config.Path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if !info.Mode().IsRegular() || info.Size() == 0 {
		return nil, ErrUncommissioned
	}

	db, err := openFile(config.Path)
	if err != nil {
		return nil, err
	}

	pool, err := db.DB()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err = checkCommissioned(ctx, pool); err == nil {
		err = apply(ctx, pool)
	}
	if err == nil {
		err = Check(ctx, pool)
	}
	if err != nil {
		return nil, errors.Join(err, pool.Close())
	}

	return db, nil
}

func Initialize(ctx context.Context, path string) (err error) {
	if err = RequireVacant(path); err != nil {
		return err
	}

	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		return err
	}

	if err = file.Close(); err != nil {
		return err
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, os.Remove(path))
		}
	}()

	db, err := openFile(path)
	if err != nil {
		return err
	}

	pool, err := db.DB()
	if err != nil {
		return err
	}

	defer func() { err = errors.Join(err, pool.Close()) }()

	if err = apply(ctx, pool); err != nil {
		return err
	}

	if err = Check(ctx, pool); err != nil {
		return err
	}

	_, err = pool.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)")
	return err
}

func RequireVacant(path string) error {
	if path == "" {
		return errors.New("database path is required")
	}

	for _, existing := range []string{path, path + "-wal", path + "-shm", filepath.Join(filepath.Dir(path), "."+filepath.Base(path)+"-litestream")} {
		if _, err := os.Lstat(existing); err == nil {
			return fmt.Errorf("retained database state: %w", os.ErrExist)
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	return nil
}

func openFile(path string) (*gorm.DB, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	uri := url.URL{Scheme: "file", Path: absolute}
	query := url.Values{"mode": {"rw"}, "_pragma": {"foreign_keys(1)", "journal_mode(WAL)", "synchronous(FULL)", "busy_timeout(1000)"}, "_txlock": {"immediate"}}
	uri.RawQuery = query.Encode()
	pool, err := sql.Open("sqlite", uri.String())
	if err != nil {
		return nil, err
	}

	pool.SetMaxOpenConns(1)
	pool.SetMaxIdleConns(1)

	db, err := gorm.Open(driver.New(driver.Config{DriverName: "sqlite", Conn: pool}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, errors.Join(err, pool.Close())
	}

	return db, nil
}

func NewSQLDB(db *gorm.DB) (*sql.DB, error) { return db.DB() }

func CheckFile(ctx context.Context, path string) error {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	uri := url.URL{Scheme: "file", Path: absolute}
	uri.RawQuery = "mode=ro"
	db, err := sql.Open("sqlite", uri.String())
	if err != nil {
		return err
	}

	defer func() { _ = db.Close() }()
	db.SetMaxOpenConns(1)

	return Check(ctx, db)
}

func HookConnection(lifecycle fx.Lifecycle, db *sql.DB) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error { return db.PingContext(ctx) },
		OnStop:  func(context.Context) error { return db.Close() },
	})
}

type HealthChecker struct{ client *sql.DB }

func NewHealthChecker(client *sql.DB) *HealthChecker { return &HealthChecker{client: client} }
func (checker *HealthChecker) CheckHealth(ctx context.Context) error {
	return checker.client.PingContext(ctx)
}
