package main

import (
	"astro/config"
	"astro/habit"
	"astro/health"
	"astro/http"
	"astro/logger"
	"astro/metrics"
	"astro/postgres"
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
