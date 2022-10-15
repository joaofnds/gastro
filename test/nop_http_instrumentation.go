package test

import (
	astrofiber "astro/http/fiber"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
)

var NopHTTPInstrumentation = fx.Decorate(newNopHTTPInstrumentation)

type nopHTTPInstrumentation struct{}

type PromHabitInstrumentation struct{}

func newNopHTTPInstrumentation() astrofiber.Instrumentation {
	return &nopHTTPInstrumentation{}
}

func (i *nopHTTPInstrumentation) Middleware(ctx *fiber.Ctx) error {
	return ctx.Next()
}
