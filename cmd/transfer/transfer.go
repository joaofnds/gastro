package main

import (
	"astro/internal/transfer"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

type commandOptions struct {
	Direction, Output, SQLitePath, TargetSchema, TargetDatabase string
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	os.Exit(run(ctx, os.Args[1:], os.Getenv, os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("transfer", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	direction := flags.String("direction", "", "postgres-to-sqlite, sqlite-to-postgres, or verify")
	output := flags.String("output", "", "new SQLite output path, must not exist")
	sqlitePath := flags.String("sqlite-path", "", "existing SQLite file")
	schema := flags.String("postgres-schema", "public", "Postgres source schema for transfer or verify")
	targetSchema := flags.String("target-schema", "", "new astro_return_<suffix> return schema")
	targetDatabase := flags.String("target-database", "", "exact Postgres return database confirmation")
	sourceFrozen := flags.Bool("source-frozen", false, "acknowledge all source writers have been stopped externally")
	timeout := flags.Duration("timeout", 5*time.Minute, "maximum operation duration")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			flags.SetOutput(stdout)
			flags.PrintDefaults()
			return 0
		}

		_, _ = fmt.Fprintln(stderr, "invalid flags; use --help")
		return 2
	}

	options := commandOptions{Direction: *direction, Output: *output, SQLitePath: *sqlitePath, TargetSchema: *targetSchema, TargetDatabase: *targetDatabase}
	if flags.NArg() != 0 || *timeout <= 0 || !validDirection(options) {
		_, _ = fmt.Fprintln(stderr, "invalid transfer options; use --help")
		return 2
	}

	dsn := getenv("ASTRO_TRANSFER_POSTGRES_DSN")
	if dsn == "" {
		_, _ = fmt.Fprintln(stderr, "ASTRO_TRANSFER_POSTGRES_DSN is required")
		return 2
	}

	ctx, cancel := context.WithTimeout(ctx, *timeout)
	defer cancel()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, transfer.ErrUnavailable.Error())
		return 1
	}

	defer func() { _ = db.Close() }()
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	var summary transfer.Summary
	switch *direction {
	case "postgres-to-sqlite":
		summary, err = transfer.ToSQLite(ctx, transfer.ToSQLiteOptions{Postgres: db, PostgresSchema: *schema, OutputPath: *output, SourceFrozen: *sourceFrozen})
	case "sqlite-to-postgres":
		summary, err = transfer.ToPostgres(ctx, transfer.ToPostgresOptions{Postgres: db, SQLitePath: *sqlitePath, TargetSchema: *targetSchema, TargetDatabase: *targetDatabase, SourceFrozen: *sourceFrozen})
	case "verify":
		summary, err = transfer.Verify(ctx, transfer.VerifyOptions{Postgres: db, PostgresSchema: *schema, SQLitePath: *sqlitePath})
	}
	if err != nil {
		if ctx.Err() != nil {
			err = ctx.Err()
		}

		_, _ = fmt.Fprintln(stderr, failureMessage(err))
		return 1
	}

	result := struct {
		Direction string           `json:"direction"`
		Verified  bool             `json:"verified"`
		Tables    transfer.Summary `json:"tables"`
	}{Direction: *direction, Verified: true, Tables: summary}
	if err := json.NewEncoder(stdout).Encode(result); err != nil {
		_, _ = fmt.Fprintln(stderr, "could not write transfer summary")
		return 1
	}

	return 0
}

func validDirection(options commandOptions) bool {
	switch options.Direction {
	case "postgres-to-sqlite":
		return options.Output != "" && options.SQLitePath == "" && options.TargetSchema == "" && options.TargetDatabase == ""
	case "sqlite-to-postgres":
		return options.Output == "" && options.SQLitePath != "" && options.TargetSchema != "" && options.TargetDatabase != ""
	case "verify":
		return options.Output == "" && options.SQLitePath != "" && options.TargetSchema == "" && options.TargetDatabase == ""
	default:
		return false
	}
}

func failureMessage(err error) string {
	for _, publicError := range []error{
		transfer.ErrTargetExists, transfer.ErrSourceNotFrozen, transfer.ErrUnsafeTarget,
		transfer.ErrInvalidOptions, transfer.ErrInvalidData, transfer.ErrRead,
		transfer.ErrWrite, transfer.ErrValidation, transfer.ErrMismatch, transfer.ErrIO,
		transfer.ErrUnavailable, context.Canceled, context.DeadlineExceeded,
	} {
		if errors.Is(err, publicError) {
			return publicError.Error()
		}
	}

	return "transfer did not complete"
}
