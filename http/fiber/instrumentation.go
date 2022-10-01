package fiber

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type PromHTTPInstrumentation struct {
	req *prometheus.CounterVec
}

type PromHabitInstrumentation struct{}

func NewPromHTTPInstrumentation() HTTPInstrumentation {
	return &PromHTTPInstrumentation{
		req: promauto.NewCounterVec(
			prometheus.CounterOpts{Name: "astro_request"},
			[]string{"method", "path"},
		),
	}
}

func (i *PromHTTPInstrumentation) Middleware(ctx *fiber.Ctx) error {
	defer i.req.WithLabelValues(ctx.Method(), ctx.Path()).Inc()
	return ctx.Next()
}
