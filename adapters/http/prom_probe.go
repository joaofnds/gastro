package http

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	lblIP     = "ip"
	lblMethod = "method"
	lblPath   = "path"
	lblStatus = "status"
)

type Probe interface {
	Middleware(*fiber.Ctx) error
}

type PromProbe struct {
	req *prometheus.CounterVec
}

func NewPromProbe() *PromProbe {
	return &PromProbe{
		req: promauto.NewCounterVec(
			prometheus.CounterOpts{Name: "astro_request"},
			[]string{lblIP, lblMethod, lblPath, lblStatus},
		),
	}
}

func (i *PromProbe) Middleware(ctx *fiber.Ctx) error {
	defer i.LogReq(ctx)
	return ctx.Next()
}

func (i *PromProbe) LogReq(ctx *fiber.Ctx) {
	labels := prometheus.Labels{}
	labels[lblIP] = ctx.Get("Fly-Client-IP", ctx.IP())
	labels[lblMethod] = ctx.Route().Method
	labels[lblPath] = ctx.Route().Path
	labels[lblStatus] = strconv.Itoa(ctx.Response().StatusCode())

	i.req.With(labels).Inc()
}
