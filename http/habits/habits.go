package habits

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
)

var Providers = fx.Options(fx.Provide(NewHabitsController), fx.Invoke(HabitsHandler))

func HabitsHandler(app *fiber.App, controller *habitsController) {
	controller.RegisterHandlers(app)
}
