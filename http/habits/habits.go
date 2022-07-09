package habits

import (
	"astro/habit"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var (
	Providers = fx.Invoke(HealthHandler)
)

func HealthHandler(app *fiber.App, svc *habit.HabitService, logger *zap.Logger) {
	app.Get("/habits", func(c *fiber.Ctx) error {
		habits, err := svc.List(c.Context())

		if err != nil {
      logger.Error("failed to list habits", zap.Error(err))
			return c.SendStatus(http.StatusInternalServerError)
		}

		return c.Status(http.StatusOK).JSON(habits)
	})

	app.Post("/habits", func(c *fiber.Ctx) error {
		name := c.Query("name")
		if name == "" {
			return c.SendStatus(http.StatusBadRequest)
		}
		_, err := svc.Create(c.Context(), name)

		if err != nil {
      logger.Error("failed to create habit", zap.Error(err))
			return c.SendStatus(http.StatusInternalServerError)
		}

		return c.SendStatus(http.StatusCreated)
	})
}
