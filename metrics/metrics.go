package metrics

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module(
	"metrics",
	fx.Provide(NewServer),
	fx.Invoke(HookHandler),
)

type Server = http.Server

func NewServer(config Config) *Server {
	http.Handle("/metrics", promhttp.Handler())
	return &http.Server{Addr: config.Address}
}

func HookHandler(lc fx.Lifecycle, server *Server, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				err := server.ListenAndServe()
				if err != nil {
					logger.Fatal("failed to start metrics server", zap.Error(err))
				}
			}()
			logger.Info("metrics server started")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})
}
