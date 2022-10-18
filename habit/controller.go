package habit

import (
	"astro/token"
	"errors"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func NewController(
	habitService *Service,
	tokenService *token.Service,
	logger *zap.Logger,
) *Controller {
	return &Controller{
		habitService: habitService,
		tokenService: tokenService,
		logger:       logger,
	}
}

type Controller struct {
	habitService *Service
	tokenService *token.Service
	logger       *zap.Logger
}

func (c Controller) Register(app *fiber.App) {
	habits := app.Group("/habits", c.middlewareDecodeToken)
	habits.Get("/", c.list)
	habits.Post("/", c.create)
	habits.Delete("/:habitID/:activityID", c.deleteActivity)

	habit := habits.Group("/:id", c.middlewareFindHabit)
	habit.Get("/", c.get)
	habit.Post("/", c.addActivity)
	habit.Delete("/", c.delete)
}

func (c Controller) list(ctx *fiber.Ctx) error {
	userID := ctx.Locals("userID").(string)
	habits, err := c.habitService.List(ctx.Context(), userID)

	if err != nil {
		c.logger.Error("failed to list habits", zap.Error(err))
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	return ctx.Status(http.StatusOK).JSON(habits)
}

func (Controller) get(ctx *fiber.Ctx) error {
	h := ctx.Locals("habit")
	return ctx.Status(http.StatusOK).JSON(h)
}

func (c Controller) delete(ctx *fiber.Ctx) error {
	userID := ctx.Locals("userID").(string)
	h := ctx.Locals("habit").(Habit)

	err := c.habitService.Delete(ctx.Context(), FindDTO{HabitID: h.ID, UserID: userID})
	if err != nil {
		c.logger.Error("failed to delete", zap.Error(err))
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	return ctx.SendStatus(http.StatusOK)
}

func (c Controller) deleteActivity(ctx *fiber.Ctx) error {
	habitID := ctx.Params("habitID")
	activityID := ctx.Params("activityID")
	userID := ctx.Locals("userID").(string)

	find := FindActivityDTO{HabitID: habitID, ActivityID: activityID, UserID: userID}
	activity, err := c.habitService.FindActivity(ctx.Context(), find)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ctx.SendStatus(http.StatusNotFound)
		}
		return ctx.SendStatus(http.StatusBadRequest)
	}

	err = c.habitService.DeleteActivity(ctx.Context(), activity)
	if err != nil {
		c.logger.Error("failed to delete activity", zap.Error(err))
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	return ctx.SendStatus(http.StatusOK)
}

func (c Controller) addActivity(ctx *fiber.Ctx) error {
	h := ctx.Locals("habit").(Habit)

	_, err := c.habitService.AddActivity(ctx.Context(), h, time.Now().UTC())
	if err != nil {
		c.logger.Error("failed to add activity", zap.Error(err))
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	return ctx.SendStatus(http.StatusCreated)
}

func (c Controller) create(ctx *fiber.Ctx) error {
	name := ctx.Query("name")
	if name == "" {
		return ctx.SendStatus(http.StatusBadRequest)
	}

	userID := ctx.Locals("userID").(string)
	h, err := c.habitService.Create(ctx.Context(), CreateDTO{UserID: userID, Name: name})
	if err != nil {
		c.logger.Error("failed to create habit", zap.Error(err))
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	return ctx.Status(http.StatusCreated).JSON(h)
}

func (c Controller) middlewareFindHabit(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if !IsUUID(id) {
		return ctx.SendStatus(http.StatusNotFound)
	}

	userID := ctx.Locals("userID").(string)

	h, err := c.habitService.Find(ctx.Context(), FindDTO{HabitID: id, UserID: userID})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ctx.SendStatus(http.StatusNotFound)
		}

		c.logger.Error("failed to get habit", zap.Error(err))
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	ctx.Locals("habit", h)

	return ctx.Next()
}

func (c Controller) middlewareDecodeToken(ctx *fiber.Ctx) error {
	token, ok := ctx.GetReqHeaders()["Authorization"]
	if !ok {
		return ctx.Status(http.StatusUnauthorized).SendString("missing Authorization token")
	}

	id, err := c.tokenService.IDFromToken([]byte(token))
	if err != nil {
		return ctx.SendStatus(http.StatusUnauthorized)
	}

	ctx.Locals("userID", string(id))
	return ctx.Next()
}
