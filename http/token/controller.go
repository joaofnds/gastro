package token

import (
	"astro/token"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func NewTokenController(service *token.TokenService) *TokenController {
	return &TokenController{service}
}

type TokenController struct {
	service *token.TokenService
}

func (c *TokenController) Register(app *fiber.App) {
	app.Post("/token", c.Create)
	app.Get("/tokentest", c.TestToken)
}

func (c *TokenController) Create(ctx *fiber.Ctx) error {
	token, err := c.service.NewToken()
	if err != nil {
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	return ctx.Status(http.StatusCreated).Send(token)
}

func (c *TokenController) TestToken(ctx *fiber.Ctx) error {
	headers := ctx.GetReqHeaders()
	token, ok := headers["Authorization"]
	if !ok {
		return ctx.Status(http.StatusBadRequest).SendString("missing Authorization header")
	}

	_, err := c.service.IDFromToken([]byte(token))
	if err != nil {
		return ctx.SendStatus(http.StatusBadRequest)
	}

	return ctx.SendStatus(http.StatusOK)
}
