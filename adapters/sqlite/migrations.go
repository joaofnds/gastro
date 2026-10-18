package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
)

const initialVersion int64 = 20261018000001
const currentVersion int64 = initialVersion

//go:embed migrations/*.sql
var migrations embed.FS

func apply(ctx context.Context, db *sql.DB) error {
	files, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return err
	}

	provider, err := goose.NewProvider(goose.DialectSQLite3, db, files)
	if err != nil {
		return err
	}

	_, err = provider.Up(ctx)
	return err
}

func checkSchema(ctx context.Context, db *sql.DB) error {
	var version int64
	if err := db.QueryRowContext(ctx, "SELECT max(version_id) FROM goose_db_version WHERE is_applied=1").Scan(&version); err != nil {
		return fmt.Errorf("%w: schema version unavailable", ErrInvalidDatabase)
	}

	if version < initialVersion || version > currentVersion {
		return fmt.Errorf("%w: unsupported schema version %d", ErrInvalidDatabase, version)
	}

	statements := []string{
		"SELECT id,user_id,name FROM habits LIMIT 0",
		"SELECT id,habit_id,description,created_at FROM activities LIMIT 0",
		"SELECT id,user_id,name FROM groups LIMIT 0",
		"SELECT group_id,habit_id,user_id FROM groups_habits LIMIT 0",
	}
	for _, statement := range statements {
		rows, err := db.QueryContext(ctx, statement)
		if err != nil {
			return fmt.Errorf("%w: incomplete application schema", ErrInvalidDatabase)
		}

		if err = rows.Close(); err != nil {
			return err
		}
	}
	return nil
}

func Check(ctx context.Context, db *sql.DB) error {
	if err := checkSchema(ctx, db); err != nil {
		return err
	}

	if err := checkStructure(ctx, db); err != nil {
		return err
	}

	var result string
	if err := db.QueryRowContext(ctx, "PRAGMA quick_check").Scan(&result); err != nil {
		return err
	}

	if result != "ok" {
		return fmt.Errorf("%w: integrity check rejected database", ErrInvalidDatabase)
	}

	rows, err := db.QueryContext(ctx, "PRAGMA foreign_key_check")
	if err != nil {
		return err
	}

	defer func() { _ = rows.Close() }()
	if rows.Next() {
		return fmt.Errorf("%w: foreign key violations", ErrInvalidDatabase)
	}

	return rows.Err()
}
