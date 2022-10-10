package token

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
)

var Providers = fx.Options(
	fx.Provide(NewTokenController),
	fx.Invoke(func(app *fiber.App, controller *TokenController) {
		controller.Register(app)
	}),
)
