package main

import (
	"astro/adapters/backuphealth"
	"astro/adapters/health"
	"astro/adapters/http"
	"astro/adapters/logger"
	"astro/adapters/metrics"
	"astro/adapters/sqlite"

	"astro/config"
	"astro/habit"
	"astro/token"

	"go.uber.org/fx"
)

func main() {
	fx.New(
		config.Module,
		logger.Module,
		metrics.Module,
		backuphealth.Module,
		health.Module,

		sqlite.Module,
		http.Module,

		habit.Module,
		token.Module,
	).Run()
}
