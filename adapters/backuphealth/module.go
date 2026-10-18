package backuphealth

import (
	"astro/adapters/sqlite"
	"context"
	"net/http"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module(
	"backuphealth",
	fx.Provide(newRuntimeChecker, fx.Private),
	fx.Invoke(registerHooks),
)

func newRuntimeChecker(config Config, database sqlite.Config) (*Checker, error) {
	return NewChecker(config, Settings{
		DatabasePath: database.Path,
		SocketPath:   "/tmp/litestream.sock",
	}, &http.Transport{Proxy: http.ProxyFromEnvironment})
}

func registerHooks(lifecycle fx.Lifecycle, checker *Checker, logger *zap.Logger) {
	var cancel context.CancelFunc
	var done chan struct{}
	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			var ctx context.Context
			ctx, cancel = context.WithCancel(context.Background())
			done = make(chan struct{})
			go func() {
				defer close(done)
				checker.Run(ctx, time.Minute, logger)
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			cancel()
			select {
			case <-done:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	})
}
