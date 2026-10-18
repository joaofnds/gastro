package transfer

import (
	astrosqlite "astro/adapters/sqlite"
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
)

func publishSQLite(ctx context.Context, source snapshot, output string) (Summary, error) {
	absOutput, err := filepath.Abs(output)
	if err != nil {
		return Summary{}, ErrIO
	}

	stageDirectory, err := os.MkdirTemp(filepath.Dir(absOutput), ".astro-transfer-")
	if err != nil {
		return Summary{}, ErrIO
	}

	defer func() { _ = os.RemoveAll(stageDirectory) }()
	stagePath := filepath.Join(stageDirectory, "astro.db")
	if err := astrosqlite.Initialize(ctx, stagePath); err != nil {
		return Summary{}, ErrValidation
	}

	summary, err := populateSQLite(ctx, source, stagePath)
	if err != nil {
		return Summary{}, err
	}

	if err := ctx.Err(); err != nil {
		return Summary{}, err
	}

	if err := syncFile(stagePath); err != nil {
		return Summary{}, err
	}

	if err := astrosqlite.RequireVacant(absOutput); errors.Is(err, os.ErrExist) {
		return Summary{}, ErrTargetExists
	} else if err != nil {
		return Summary{}, ErrIO
	}

	if err := os.Link(stagePath, absOutput); errors.Is(err, os.ErrExist) {
		return Summary{}, ErrTargetExists
	} else if err != nil {
		return Summary{}, ErrIO
	}

	if err := syncDirectory(filepath.Dir(absOutput)); err != nil {
		return Summary{}, err
	}

	return summary, nil
}

func populateSQLite(ctx context.Context, source snapshot, path string) (Summary, error) {
	orm, err := astrosqlite.Open(astrosqlite.Config{Path: path})
	if err != nil {
		return Summary{}, ErrValidation
	}

	db, err := orm.DB()
	if err != nil {
		return Summary{}, ErrUnavailable
	}

	defer func() { _ = db.Close() }()

	if err := copySQLite(ctx, db, source); err != nil {
		return Summary{}, err
	}

	if err := astrosqlite.Check(ctx, db); err != nil {
		return Summary{}, ErrValidation
	}

	var busy, logFrames, checkpointed int
	if err := db.QueryRowContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)").Scan(&busy, &logFrames, &checkpointed); err != nil || busy != 0 || logFrames != checkpointed {
		return Summary{}, ErrWrite
	}

	if err := db.Close(); err != nil {
		return Summary{}, ErrWrite
	}

	restored, err := readSQLite(ctx, path)
	if err != nil {
		return Summary{}, err
	}

	return compareSnapshots(source, restored)
}

func copySQLite(ctx context.Context, db *sql.DB, source snapshot) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return ErrWrite
	}

	defer func() { _ = tx.Rollback() }()

	if err := writeSnapshot(ctx, tx, source, ""); err != nil {
		return err
	}

	restored, err := readSnapshot(ctx, tx, "")
	if err != nil {
		return err
	}

	if _, err := compareSnapshots(source, restored); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return ErrWrite
	}

	return nil
}

func syncFile(path string) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return ErrIO
	}

	defer func() { _ = file.Close() }()

	if err := file.Sync(); err != nil {
		return ErrIO
	}

	if err := file.Close(); err != nil {
		return ErrIO
	}

	return nil
}

func syncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return ErrIO
	}

	defer func() { _ = directory.Close() }()

	if err := directory.Sync(); err != nil {
		return ErrIO
	}

	return nil
}
