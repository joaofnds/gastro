package main

import (
	"astro/adapters/logger"
	"astro/adapters/postgres"
	"astro/config"
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/pressly/goose/v3"
	"go.uber.org/fx"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println(`usage:
  go run cmd/migrate/migrate.go up
  go run cmd/migrate/migrate.go down
  go run cmd/migrate/migrate.go down-to 20170506082527
  go run cmd/migrate/migrate.go status
  go run cmd/migrate/migrate.go redo
	`)
		os.Exit(1)
	}

	app := fx.New(
		logger.NopLogger,
		config.Module,
		postgres.Module,
		fx.Invoke(func(db *sql.DB, config postgres.Config) error {
			return goose.Run(os.Args[1], db, "cmd/migrate/migrations", os.Args[2:]...)
		}),
	)
	defer func() { must(app.Stop(context.Background())) }()
	must(app.Start(context.Background()))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
