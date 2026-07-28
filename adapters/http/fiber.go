package http

import (
	"context"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"go.uber.org/fx"
)

var FiberModule = fx.Module(
	"fiber",
	fx.Provide(NewFiber),
	fx.Invoke(HookFiber),
	fx.Provide(NewPromProbe),
	fx.Provide(func(probe *PromProbe) Probe { return probe }),
)

func NewFiber(config Config, probe Probe) *fiber.App {
	app := fiber.New(fiber.Config{
		JSONEncoder:  sonic.Marshal,
		JSONDecoder:  sonic.Unmarshal,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	})
	app.Use(recover.New())
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))
	app.Use(limiter.New(limiter.Config{
		Max:               config.Limiter.Requests,
		Expiration:        config.Limiter.Expiration,
		LimiterMiddleware: limiter.SlidingWindow{},
	}))
	app.Use(probe.Middleware)
	// v3's default AllowMethods adds QUERY, which the service defines no handler for.
	// Listing v2's six keeps the advertised set unchanged.
	app.Use(cors.New(cors.Config{
		AllowMethods: []string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodHead,
			fiber.MethodPut,
			fiber.MethodDelete,
			fiber.MethodPatch,
		},
	}))
	return app
}

func HookFiber(lc fx.Lifecycle, app *fiber.App, config Config) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				listenConfig := fiber.ListenConfig{DisableStartupMessage: true}
				if err := app.Listen(fmt.Sprintf(":%d", config.Port), listenConfig); err != nil {
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
