package fiber

import (
	"astro/config"
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
)

var (
	Module = fx.Options(
		fx.Provide(NewFiber),
		fx.Invoke(HookFiber),
	)
)

func NewFiber() *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		return c.Next()
	})
	return app
}

func HookFiber(lc fx.Lifecycle, app *fiber.App, config config.AppConfig) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				if err := app.Listen(fmt.Sprintf(":%d", config.Port)); err != nil {
					panic(err)
				}
			}()
			return nil
		},
		OnStop: func(context.Context) error {
			return app.Shutdown()
		},
	})
}
