package http

import (
	"astro/http/fiber"
	"astro/http/habits"
	"astro/http/health"

	"go.uber.org/fx"
)

var (
	Module = fx.Options(
		fiber.Module,
		health.Providers,
		habits.Providers,
	)
)
