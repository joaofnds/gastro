package token

import (
	"astro/token"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
)

var Providers = fx.Invoke(TokenHandler)

func TokenHandler(app *fiber.App, tokenService *token.TokenService) {
	app.Post("/token", func(c *fiber.Ctx) error {
		token, err := tokenService.NewToken()
		if err != nil {
			return c.SendStatus(http.StatusInternalServerError)
		}

		return c.Status(http.StatusCreated).Send(token)
	})

	app.Get("/tokentest", func(c *fiber.Ctx) error {
		headers := c.GetReqHeaders()
		token, ok := headers["Authorization"]
		if !ok {
			return c.Status(http.StatusBadRequest).SendString("missing Authorization header")
		}
		_, err := tokenService.IdFromToken([]byte(token))
		if err != nil {
			return c.SendStatus(http.StatusBadRequest)
		}

		return c.SendStatus(http.StatusOK)
	})
}
