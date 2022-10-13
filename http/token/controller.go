package token

import (
	"astro/token"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func NewController(service *token.Service) *Controller {
	return &Controller{service}
}

type Controller struct {
	service *token.Service
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
	headers := ctx.GetReqHeaders()
	tok, ok := headers["Authorization"]
	if !ok {
		return ctx.Status(http.StatusBadRequest).SendString("missing Authorization header")
	}

	_, err := c.service.IDFromToken([]byte(tok))
	if err != nil {
		return ctx.SendStatus(http.StatusBadRequest)
	}

	return ctx.SendStatus(http.StatusOK)
}
