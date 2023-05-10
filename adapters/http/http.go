package http

import (
	"astro/adapters/health"
	"astro/habit"
	"astro/token"

	"github.com/gofiber/fiber/v2"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"http",
	FiberModule,
	fx.Invoke(func(
		app *fiber.App,
		healthController *health.Controller,
		habitController *habit.Controller,
		tokenController *token.Controller,
	) {
		app.Get("/", func(c *fiber.Ctx) error {
			return c.Redirect("https://github.com/joaofnds/astro")
		})
		healthController.Register(app)
		habitController.Register(app)
		tokenController.Register(app)
	}),
)
