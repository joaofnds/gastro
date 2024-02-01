package token

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func NewController(service *Service) *Controller {
	return &Controller{service}
}

type Controller struct {
	service *Service
}

func (c *Controller) Register(app *fiber.App) {
	app.Post("/token", c.Create)
	app.Get("/tokentest", c.TestToken)
}

func (c *Controller) Create(ctx *fiber.Ctx) error {
	tok, err := c.service.NewToken()
	if err != nil {
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	return ctx.Status(http.StatusCreated).Send(tok)
}

func (c *Controller) TestToken(ctx *fiber.Ctx) error {
	tok := ctx.Get("Authorization")
	if tok == "" {
		return ctx.Status(http.StatusBadRequest).SendString("missing Authorization header")
	}

	_, err := c.service.IDFromToken([]byte(tok))
	if err != nil {
		return ctx.SendStatus(http.StatusBadRequest)
	}

	return ctx.SendStatus(http.StatusOK)
}
