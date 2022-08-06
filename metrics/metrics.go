package metrics

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Options(
	fx.Provide(NewMetricsServer),
	fx.Invoke(HookMetricsHandler),
)

type MetricsServer = http.Server

func NewMetricsServer() *MetricsServer {
	http.Handle("/metrics", promhttp.Handler())
	server := http.Server{Addr: ":9091"}
	return &server
}

func HookMetricsHandler(lc fx.Lifecycle, server *MetricsServer, logger *zap.Logger) {
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
