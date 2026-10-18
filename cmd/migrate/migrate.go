package main

import (
	"astro/adapters/sqlite"
	"context"
	"errors"
	"fmt"
	"os"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() (err error) {
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: migrate init|up|check (database path from SQLITE_PATH)")
	}

	path := os.Getenv("SQLITE_PATH")
	if path == "" {
		path = "./data/astro.db"
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	switch os.Args[1] {
	case "init":
		return sqlite.Initialize(ctx, path)
	case "check":
		return sqlite.CheckFile(ctx, path)
	case "up":
		db, openErr := sqlite.Open(sqlite.Config{Path: path})
		if openErr != nil {
			return openErr
		}

		pool, poolErr := db.DB()
		if poolErr != nil {
			return poolErr
		}

		defer func() { err = errors.Join(err, pool.Close()) }()
		return sqlite.Check(ctx, pool)
	default:
		return fmt.Errorf("unknown migration operation %q", os.Args[1])
	}
}
