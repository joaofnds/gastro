package health

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
)

var Providers = fx.Options(
	fx.Provide(NewController),
)

type Controller struct {
	service Checker
}

func NewController(service Checker) *Controller {
	return &Controller{service}
}

func (c *Controller) Register(app *fiber.App) {
	app.Get("/health", c.CheckHealth)
}

func (c *Controller) CheckHealth(ctx *fiber.Ctx) error {
	status := http.StatusOK

	h := c.service.CheckHealth()
	if !h.AllUp() {
		status = http.StatusServiceUnavailable
	}

	return ctx.Status(status).JSON(h)
}
