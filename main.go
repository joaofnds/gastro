package main

import (
	"astro/adapters/health"
	"astro/adapters/http"
	"astro/adapters/logger"
	"astro/adapters/metrics"
	"astro/adapters/postgres"
	"astro/config"
	"astro/habit"
	"astro/token"

	"go.uber.org/fx"
)

func main() {
	fx.New(
		logger.Module,
		config.Module,
		metrics.Module,
		health.Module,
		http.Module,
		postgres.Module,
		habit.Module,
		token.Module,
	).Run()
}
