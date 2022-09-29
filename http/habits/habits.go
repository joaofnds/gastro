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

	habits := app.Group("/habits", c.middlewareDecodeToken)
	habits.Get("/", c.list)
	habits.Post("/", c.create)

	habit := habits.Group("/:name", c.middlewareFindHabit)
	habit.Get("/", c.get)
	habit.Post("/", c.addActivity)
	habit.Delete("/", c.delete)
}
