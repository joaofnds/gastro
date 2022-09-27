package habits

import (
	"astro/habit"
	"astro/token"
	"errors"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type habitsController struct {
	habitService *habit.HabitService
	tokenService *token.TokenService
	logger       *zap.Logger
}

func (c habitsController) list(ctx *fiber.Ctx) error {
	habits, err := c.habitService.List(ctx.Context())

	if err != nil {
		c.logger.Error("failed to list habits", zap.Error(err))
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	return ctx.Status(http.StatusOK).JSON(habits)
}

func (habitsController) get(ctx *fiber.Ctx) error {
	h := ctx.Locals("habit")
	return ctx.Status(http.StatusOK).JSON(h)
}

func (c habitsController) delete(ctx *fiber.Ctx) error {
	h := ctx.Locals("habit").(habit.Habit)

	err := c.habitService.DeleteByName(ctx.Context(), h.Name)
	if err != nil {
		c.logger.Error("failed to delete habit by name", zap.Error(err))
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	return ctx.SendStatus(http.StatusOK)
}

func (c habitsController) addActivity(ctx *fiber.Ctx) error {
	h := ctx.Locals("habit").(habit.Habit)

	_, err := c.habitService.AddActivity(ctx.Context(), h, time.Now().UTC())
	if err != nil {
		c.logger.Error("failed to add activity", zap.Error(err))
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	return ctx.SendStatus(http.StatusCreated)
}

func (c habitsController) create(ctx *fiber.Ctx) error {
	name := ctx.Query("name")
	if name == "" {
		return ctx.SendStatus(http.StatusBadRequest)
	}

	h, err := c.habitService.Create(ctx.Context(), name)
	if err != nil {
		c.logger.Error("failed to create habit", zap.Error(err))
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	return ctx.Status(http.StatusCreated).JSON(h)
}

func (c habitsController) middlewareFindHabit(ctx *fiber.Ctx) error {
	name := ctx.Params("name")
	if name == "" {
		return ctx.SendStatus(http.StatusBadRequest)
	}

	h, err := c.habitService.FindByName(ctx.Context(), name)
	if err != nil {
		if errors.Is(err, habit.HabitNotFoundErr) {
			return ctx.SendStatus(http.StatusNotFound)
		}

		c.logger.Error("failed to get habit", zap.Error(err))
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	ctx.Locals("habit", h)

	return ctx.Next()
}

func (c habitsController) middlewareDecodeToken(ctx *fiber.Ctx) error {
	token, ok := ctx.GetReqHeaders()["Authorization"]
	if !ok {
		return ctx.Status(http.StatusUnauthorized).SendString("missing Authorization token")
	}

	id, err := c.tokenService.IdFromToken([]byte(token))
	if err != nil {
		return ctx.SendStatus(http.StatusUnauthorized)
	}

	ctx.Locals("userID", string(id))
	return ctx.Next()
}
