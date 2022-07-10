package habits

import (
	"astro/habit"
	"errors"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var (
	Providers = fx.Invoke(HealthHandler)
)

func HealthHandler(app *fiber.App, habitService *habit.HabitService, logger *zap.Logger) {
	findHabitByName := func(c *fiber.Ctx) error {
		name := c.Params("name")
		if name == "" {
			return c.SendStatus(http.StatusBadRequest)
		}

		h, err := habitService.FindByName(c.Context(), name)
		if err != nil {
			if errors.Is(err, habit.HabitNotFoundErr) {
				return c.SendStatus(http.StatusNotFound)
			}

			logger.Error("failed to get habit", zap.Error(err))
			return c.SendStatus(http.StatusInternalServerError)
		}

		c.Locals("habit", h)

		return c.Next()
	}

	app.Get("/habits", func(c *fiber.Ctx) error {
		habits, err := habitService.List(c.Context())

		if err != nil {
			logger.Error("failed to list habits", zap.Error(err))
			return c.SendStatus(http.StatusInternalServerError)
		}

		return c.Status(http.StatusOK).JSON(habits)
	})

	app.Get("/habits/:name", findHabitByName, func(c *fiber.Ctx) error {
		h := c.Locals("habit")
		return c.Status(http.StatusOK).JSON(h)
	})

	app.Post("/habits/:name", findHabitByName, func(c *fiber.Ctx) error {
		h := c.Locals("habit").(habit.Habit)

		_, err := habitService.AddActivity(c.Context(), h, time.Now().UTC())
		if err != nil {
			logger.Error("failed to add activity", zap.Error(err))
			return c.SendStatus(http.StatusInternalServerError)
		}

		return c.SendStatus(http.StatusCreated)
	})

	app.Post("/habits", func(c *fiber.Ctx) error {
		name := c.Query("name")
		if name == "" {
			return c.SendStatus(http.StatusBadRequest)
		}
		_, err := habitService.Create(c.Context(), name)

		if err != nil {
			logger.Error("failed to create habit", zap.Error(err))
			return c.SendStatus(http.StatusInternalServerError)
		}

		return c.SendStatus(http.StatusCreated)
	})
}
