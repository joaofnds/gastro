package http

import (
	"astro/http/fiber"
	"astro/http/habits"
	"astro/http/health"
	"astro/http/token"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"http",
	fiber.Module,
	health.Providers,
	habits.Providers,
	token.Providers,
)
