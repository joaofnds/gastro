package fiber

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	lblMethod = "method"
	lblPath   = "path"
	lblStatus = "status"
)

type PromHTTPInstrumentation struct {
	req *prometheus.CounterVec
}

type PromHabitInstrumentation struct{}

func NewPromHTTPInstrumentation() HTTPInstrumentation {
	return &PromHTTPInstrumentation{
		req: promauto.NewCounterVec(
			prometheus.CounterOpts{Name: "astro_request"},
			[]string{lblMethod, lblPath, lblStatus},
		),
	}
}

func (i *PromHTTPInstrumentation) Middleware(ctx *fiber.Ctx) error {
	defer i.LogReq(ctx)
	return ctx.Next()
}

func (i *PromHTTPInstrumentation) LogReq(ctx *fiber.Ctx) {
	req, res := ctx.Request(), ctx.Response()

	labels := prometheus.Labels{}
	labels[lblMethod] = string(req.Header.Method())
	labels[lblPath] = string(req.URI().Path())
	labels[lblStatus] = strconv.Itoa(res.StatusCode())

	i.req.With(labels).Inc()
}
