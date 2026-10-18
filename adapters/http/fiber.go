package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"go.uber.org/fx"
)

const repositoryURL = "https://github.com/joaofnds/astro"

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
	app.Get("/", func(c fiber.Ctx) error {
		return c.Redirect().Status(fiber.StatusFound).To(repositoryURL)
	})
	return app
}

func HookFiber(lifecycle fx.Lifecycle, app *fiber.App, config Config, shutdowner fx.Shutdowner) {
	var listener *startupListener
	stopped := make(chan error, 1)
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			network := net.ListenConfig{}
			bound, err := network.Listen(ctx, "tcp4", fmt.Sprintf(":%d", config.Port))
			if err != nil {
				return err
			}

			listener = &startupListener{Listener: bound, ready: make(chan struct{})}

			go func() {
				err := app.Listener(listener, fiber.ListenConfig{DisableStartupMessage: true})
				stopped <- err
				if err != nil {
					_ = shutdowner.Shutdown(fx.ExitCode(1))
				}
			}()

			select {
			case <-listener.ready:
				return nil
			case err := <-stopped:
				return errors.Join(err, listener.Close())
			case <-ctx.Done():
				closeErr := listener.Close()
				return errors.Join(ctx.Err(), closeErr, <-stopped)
			}
		},
		OnStop: func(ctx context.Context) error {
			err := app.ShutdownWithContext(ctx)
			select {
			case serveErr := <-stopped:
				return errors.Join(err, serveErr)
			case <-ctx.Done():
				return errors.Join(err, ctx.Err())
			}
		},
	})
}

type startupListener struct {
	net.Listener
	ready chan struct{}
	once  sync.Once
}

func (listener *startupListener) Accept() (net.Conn, error) {
	listener.once.Do(func() { close(listener.ready) })
	return listener.Listener.Accept()
}
