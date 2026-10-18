package transfer

import (
	astrosqlite "astro/adapters/sqlite"
	"context"
	"database/sql"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lib/pq"
)

var (
	ErrTargetExists    = errors.New("target already exists; choose a new output path or return schema")
	ErrSourceNotFrozen = errors.New("source writes must be frozen before transfer")
	ErrUnsafeTarget    = errors.New("return target requires a new astro_return_<suffix> schema and matching database confirmation")
	ErrInvalidOptions  = errors.New("invalid transfer options")
	ErrInvalidData     = errors.New("source contains unsupported UUID or timestamp values")
	ErrRead            = errors.New("could not read a consistent source snapshot")
	ErrWrite           = errors.New("could not write the isolated target")
	ErrValidation      = errors.New("database schema or integrity validation failed")
	ErrMismatch        = errors.New("field hashes or row counts differ")
	ErrIO              = errors.New("could not safely publish the database file")
	ErrUnavailable     = errors.New("database connection unavailable")
)

var schemaPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

type ToSQLiteOptions struct {
	Postgres       *sql.DB
	PostgresSchema string
	OutputPath     string
	SourceFrozen   bool
}

type ToPostgresOptions struct {
	Postgres       *sql.DB
	SQLitePath     string
	TargetSchema   string
	TargetDatabase string
	SourceFrozen   bool
}

type VerifyOptions struct {
	Postgres       *sql.DB
	PostgresSchema string
	SQLitePath     string
}

func ToSQLite(ctx context.Context, options ToSQLiteOptions) (Summary, error) {
	if !options.SourceFrozen {
		return Summary{}, ErrSourceNotFrozen
	}

	if options.OutputPath == "" || !schemaPattern.MatchString(options.PostgresSchema) {
		return Summary{}, ErrInvalidOptions
	}

	if err := astrosqlite.RequireVacant(options.OutputPath); errors.Is(err, os.ErrExist) {
		return Summary{}, ErrTargetExists
	} else if err != nil {
		return Summary{}, ErrIO
	}

	if options.Postgres == nil {
		return Summary{}, ErrUnavailable
	}

	source, err := readPostgres(ctx, options.Postgres, options.PostgresSchema)
	if err != nil {
		return Summary{}, err
	}

	return publishSQLite(ctx, source, options.OutputPath)
}

func ToPostgres(ctx context.Context, options ToPostgresOptions) (Summary, error) {
	if !safeReturnSchema(options.TargetSchema) || options.TargetDatabase == "" {
		return Summary{}, ErrUnsafeTarget
	}

	if !options.SourceFrozen {
		return Summary{}, ErrSourceNotFrozen
	}

	if options.Postgres == nil {
		return Summary{}, ErrUnavailable
	}

	source, err := readSQLite(ctx, options.SQLitePath)
	if err != nil {
		return Summary{}, err
	}

	target, err := options.Postgres.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return Summary{}, ErrUnavailable
	}

	defer func() { _ = target.Rollback() }()

	var database string
	if err := target.QueryRowContext(ctx, "SELECT current_database()").Scan(&database); err != nil {
		return Summary{}, ErrRead
	}

	if database != options.TargetDatabase {
		return Summary{}, ErrUnsafeTarget
	}

	if err := createReturnSchema(ctx, target, options.TargetSchema); err != nil {
		return Summary{}, err
	}

	if err := writeSnapshot(ctx, target, source, options.TargetSchema); err != nil {
		return Summary{}, err
	}

	restored, err := readSnapshot(ctx, target, options.TargetSchema)
	if err != nil {
		return Summary{}, err
	}

	summary, err := compareSnapshots(source, restored)
	if err != nil {
		return Summary{}, err
	}

	if err := target.Commit(); err != nil {
		return Summary{}, ErrWrite
	}

	return summary, nil
}

func Verify(ctx context.Context, options VerifyOptions) (Summary, error) {
	if !schemaPattern.MatchString(options.PostgresSchema) {
		return Summary{}, ErrInvalidOptions
	}

	if options.Postgres == nil {
		return Summary{}, ErrUnavailable
	}

	source, err := readPostgres(ctx, options.Postgres, options.PostgresSchema)
	if err != nil {
		return Summary{}, err
	}

	target, err := readSQLite(ctx, options.SQLitePath)
	if err != nil {
		return Summary{}, err
	}

	return compareSnapshots(source, target)
}

func InspectSQLite(ctx context.Context, path string) (Summary, error) {
	data, err := readSQLite(ctx, path)
	if err != nil {
		return Summary{}, err
	}

	return data.summary()
}

func readPostgres(ctx context.Context, db *sql.DB, schema string) (snapshot, error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return snapshot{}, ErrUnavailable
	}

	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, "SET LOCAL TIME ZONE 'UTC'"); err != nil {
		return snapshot{}, ErrRead
	}

	return readSnapshot(ctx, tx, schema)
}

func readSQLite(ctx context.Context, path string) (snapshot, error) {
	absPath, err := filepath.Abs(path)
	if err != nil || path == "" {
		return snapshot{}, ErrInvalidOptions
	}

	info, err := os.Stat(absPath)
	if err != nil || !info.Mode().IsRegular() {
		return snapshot{}, ErrRead
	}

	uri := &url.URL{Scheme: "file", Path: absPath}
	query := uri.Query()
	query.Set("mode", "ro")
	query.Add("_pragma", "busy_timeout(1000)")
	query.Add("_pragma", "foreign_keys(1)")
	uri.RawQuery = query.Encode()
	db, err := sql.Open("sqlite", uri.String())
	if err != nil {
		return snapshot{}, ErrUnavailable
	}

	defer func() { _ = db.Close() }()
	db.SetMaxOpenConns(1)

	if err := astrosqlite.Check(ctx, db); err != nil {
		return snapshot{}, ErrValidation
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return snapshot{}, ErrRead
	}

	defer func() { _ = tx.Rollback() }()

	var integrity string
	if err := tx.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&integrity); err != nil || integrity != "ok" {
		return snapshot{}, ErrValidation
	}

	return readSnapshot(ctx, tx, "")
}

func compareSnapshots(source, target snapshot) (Summary, error) {
	sourceSummary, err := source.summary()
	if err != nil {
		return Summary{}, err
	}

	targetSummary, err := target.summary()
	if err != nil {
		return Summary{}, err
	}

	if sourceSummary != targetSummary {
		return Summary{}, ErrMismatch
	}

	return sourceSummary, nil
}

func safeReturnSchema(schema string) bool {
	return schemaPattern.MatchString(schema) && strings.HasPrefix(schema, "astro_return_") && len(schema) > len("astro_return_")
}

func createReturnSchema(ctx context.Context, tx *sql.Tx, schema string) error {
	if _, err := tx.ExecContext(ctx, "CREATE SCHEMA "+pq.QuoteIdentifier(schema)); err != nil {
		if pgError, ok := errors.AsType[*pq.Error](err); ok && pgError.Code == "42P06" {
			return ErrTargetExists
		}

		return ErrWrite
	}

	statements := []string{
		"CREATE TABLE " + tableName(schema, "habits") + " (id uuid PRIMARY KEY DEFAULT gen_random_uuid(), user_id uuid NOT NULL, name varchar NOT NULL)",
		"CREATE UNIQUE INDEX idx_habit_id_and_user_id ON " + tableName(schema, "habits") + " (id,user_id)",
		"CREATE TABLE " + tableName(schema, "groups") + " (id uuid NOT NULL DEFAULT gen_random_uuid(), user_id uuid NOT NULL, name varchar NOT NULL, PRIMARY KEY(id,user_id))",
		"CREATE TABLE " + tableName(schema, "activities") + " (id uuid PRIMARY KEY DEFAULT gen_random_uuid(), habit_id uuid NOT NULL REFERENCES " + tableName(schema, "habits") + " (id) ON DELETE CASCADE, description varchar NOT NULL DEFAULT '', created_at timestamp NOT NULL DEFAULT now())",
		"CREATE INDEX idx_activity_habit ON " + tableName(schema, "activities") + " (habit_id)",
		"CREATE TABLE " + tableName(schema, "groups_habits") + " (group_id uuid NOT NULL, habit_id uuid NOT NULL, user_id uuid NOT NULL, PRIMARY KEY(group_id,habit_id), FOREIGN KEY(habit_id,user_id) REFERENCES " + tableName(schema, "habits") + " (id,user_id) ON DELETE CASCADE ON UPDATE CASCADE, FOREIGN KEY(group_id,user_id) REFERENCES " + tableName(schema, "groups") + " (id,user_id) ON DELETE CASCADE ON UPDATE CASCADE)",
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return ErrWrite
		}
	}

	return nil
}
