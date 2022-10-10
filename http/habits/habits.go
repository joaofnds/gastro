package habits

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
)

var Providers = fx.Options(
	fx.Provide(NewHabitsController),
	fx.Invoke(func(app *fiber.App, controller *habitsController) {
		controller.Register(app)
	}),
)
