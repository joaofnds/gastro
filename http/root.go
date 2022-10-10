package http

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
)

var RootProvider = fx.Invoke(RootHandler)

func RootHandler(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("https://github.com/joaofnds/astro")
	})
}
