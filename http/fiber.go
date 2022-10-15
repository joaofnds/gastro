package http

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/fx"
)

var FiberModule = fx.Module(
	"fiber",
	fx.Provide(NewFiber),
	fx.Invoke(HookFiber),
	fx.Provide(NewPromHTTPInstrumentation),
	fx.Provide(func(instr *PromInstrumentation) Instrumentation { return instr }),
)

type Instrumentation interface {
	Middleware(*fiber.Ctx) error
}

func NewFiber(config Config, instrumentation Instrumentation) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	app.Use(recover.New())
	app.Use(limiter.New(limiter.Config{
		Max:               config.Limiter.Requests,
		Expiration:        config.Limiter.Expiration * time.Second,
		LimiterMiddleware: limiter.SlidingWindow{},
	}))
	app.Use(instrumentation.Middleware)
	app.Use(cors.New())
	return app
}

func HookFiber(lc fx.Lifecycle, app *fiber.App, config Config) {
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