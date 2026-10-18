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
	fx.Provide(newServer, fx.Private),
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
	serverContext, stopRequests := context.WithCancel(context.Background())
	app.Hooks().OnPreShutdown(func() error {
		stopRequests()
		return nil
	})
	app.Use(func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(serverContext, 30*time.Second)
		defer cancel()

		c.SetContext(ctx)

		return c.Next()
	})
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

func newServer(app *fiber.App, config Config, shutdowner fx.Shutdowner) *server {
	return &server{
		app:     app,
		address: fmt.Sprintf(":%d", config.Port),
		shutdown: func() {
			_ = shutdowner.Shutdown(fx.ExitCode(1))
		},
		stopped: make(chan error, 1),
	}
}

func HookFiber(lifecycle fx.Lifecycle, server *server) {
	lifecycle.Append(fx.Hook{
		OnStart: server.Start,
		OnStop:  server.Stop,
	})
}
