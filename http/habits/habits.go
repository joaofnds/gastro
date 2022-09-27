package habits

import (
	"astro/habit"
	"astro/token"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Providers = fx.Invoke(HabitsHandler)

func HabitsHandler(
	app *fiber.App,
	habitService *habit.HabitService,
	tokenService *token.TokenService,
	logger *zap.Logger,
) {
	c := habitsController{
		habitService: habitService,
		tokenService: tokenService,
		logger:       logger,
	}

	app.Get("/habits", c.middlewareDecodeToken, c.list)
	app.Get("/habits/:name", c.middlewareDecodeToken, c.middlewareFindHabit, c.get)
	app.Delete("/habits/:name", c.middlewareDecodeToken, c.middlewareFindHabit, c.delete)
	app.Post("/habits/:name", c.middlewareDecodeToken, c.middlewareFindHabit, c.addActivity)
	app.Post("/habits", c.middlewareDecodeToken, c.create)
}
