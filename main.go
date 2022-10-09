package main

import (
	"astro/config"
	"astro/habit"
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
		http.Module,
		postgres.Module,
		habit.Module,
		token.Module,
	).Run()
}
