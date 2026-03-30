package metrics

import (
	"context"
	"net"
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
			ln, err := net.Listen("tcp", server.Addr)
			if err != nil {
				return err
			}
			go func() {
				if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
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
