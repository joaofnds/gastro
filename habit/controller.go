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
	groups := app.Group("/groups", c.middlewareDecodeToken)
	groups.Get("/", c.listGroupsAndHabits)
	groups.Post("/", c.createGroup)

	group := groups.Group("/:groupID", c.middlewareFindGroup)
	group.Delete("/", c.deleteGroup)

	groupHabits := group.Group("/:habitID", c.middlewareFindHabit)
	groupHabits.Post("/", c.addToGroup)
	groupHabits.Delete("/", c.removeHabitFromGroup)

	habits := app.Group("/habits", c.middlewareDecodeToken)
	habits.Get("/", c.list)
	habits.Post("/", c.create)

	habit := habits.Group("/:habitID", c.middlewareFindHabit)
	habit.Get("/", c.get)
	habit.Post("/", c.addActivity)
	habit.Patch("/", c.update)
	habit.Delete("/", c.delete)

	activity := habit.Group(":activityID", c.middlewareFindActivity)
	activity.Patch("/", c.updateActivity)
	activity.Delete("/", c.deleteActivity)
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

func (c Controller) update(ctx *fiber.Ctx) error {
	h := ctx.Locals("habit").(Habit)

	body := new(NamePayload)
	if err := ctx.BodyParser(body); err != nil {
		c.logger.Warn("failed to parse body", zap.Error(err))
		return ctx.Status(http.StatusBadRequest).SendString(err.Error())
	}

	dto := UpdateHabitDTO{Name: body.Name, HabitID: h.ID}
	if err := c.habitService.Update(ctx.Context(), dto); err != nil {
		c.logger.Error("failed to update", zap.Error(err))
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	return ctx.SendStatus(http.StatusOK)
}

func (c Controller) delete(ctx *fiber.Ctx) error {
	userID := ctx.Locals("userID").(string)
	h := ctx.Locals("habit").(Habit)

	err := c.habitService.Delete(ctx.Context(), FindHabitDTO{HabitID: h.ID, UserID: userID})
	if err != nil {
		c.logger.Error("failed to delete", zap.Error(err))
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	return ctx.SendStatus(http.StatusOK)
}

func (c Controller) addActivity(ctx *fiber.Ctx) error {
	h := ctx.Locals("habit").(Habit)

	body := new(CheckInPayload)
	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(http.StatusBadRequest).SendString(err.Error())
	}

	dto := AddActivityDTO{Desc: body.Description, Time: body.Date}
	if _, err := c.habitService.AddActivity(ctx.Context(), h, dto); err != nil {
		c.logger.Error("failed to add activity", zap.Error(err))
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	return ctx.SendStatus(http.StatusCreated)
}

func (c Controller) updateActivity(ctx *fiber.Ctx) error {
	act := ctx.Locals("activity").(Activity)

	body := new(CheckInPayload)
	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(http.StatusBadRequest).SendString(err.Error())
	}

	dto := UpdateActivityDTO{ActivityID: act.ID, Desc: body.Description}
	if _, err := c.habitService.UpdateActivity(ctx.Context(), dto); err != nil {
		c.logger.Error("failed to update activity", zap.Error(err))
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	return ctx.SendStatus(http.StatusOK)
}

func (c Controller) deleteActivity(ctx *fiber.Ctx) error {
	activity := ctx.Locals("activity").(Activity)

	if err := c.habitService.DeleteActivity(ctx.Context(), activity); err != nil {
		c.logger.Error("failed to delete activity", zap.Error(err))
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	return ctx.SendStatus(http.StatusOK)
}

func (c Controller) create(ctx *fiber.Ctx) error {
	name := ctx.Query("name")
	if name == "" {
		return ctx.SendStatus(http.StatusBadRequest)
	}

	userID := ctx.Locals("userID").(string)
	h, err := c.habitService.Create(ctx.Context(), CreateHabitDTO{UserID: userID, Name: name})
	if err != nil {
		c.logger.Error("failed to create habit", zap.Error(err))
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	return ctx.Status(http.StatusCreated).JSON(h)
}

func (c Controller) listGroupsAndHabits(ctx *fiber.Ctx) error {
	userID := ctx.Locals("userID").(string)
	groups, habits, err := c.habitService.GroupsAndHabits(ctx.Context(), userID)
	if err != nil {
		c.logger.Error("failed to list groups and habits", zap.Error(err))
	}

	return ctx.Status(http.StatusOK).JSON(GroupsAndHabitsPayload{Groups: groups, Habits: habits})
}

func (c Controller) createGroup(ctx *fiber.Ctx) error {
	body := new(NamePayload)

	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(http.StatusBadRequest).SendString(err.Error())
	}

	userID := ctx.Locals("userID").(string)
	dto := CreateGroupDTO{UserID: userID, Name: body.Name}
	group, err := c.habitService.CreateGroup(ctx.Context(), dto)
	if err != nil {
		c.logger.Error("failed to update activity", zap.Error(err))
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	return ctx.Status(http.StatusCreated).JSON(group)
}

func (c Controller) addToGroup(ctx *fiber.Ctx) error {
	group := ctx.Locals("group").(Group)
	hab := ctx.Locals("habit").(Habit)

	if err := c.habitService.AddToGroup(ctx.Context(), hab, group); err != nil {
		c.logger.Error("failed to add habit to group", zap.Error(err))
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	return ctx.SendStatus(http.StatusCreated)
}

func (c Controller) deleteGroup(ctx *fiber.Ctx) error {
	group := ctx.Locals("group").(Group)
	if err := c.habitService.DeleteGroup(ctx.Context(), group); err != nil {
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	return ctx.SendStatus(http.StatusOK)
}

func (c Controller) removeHabitFromGroup(ctx *fiber.Ctx) error {
	group := ctx.Locals("group").(Group)
	hab := ctx.Locals("habit").(Habit)

	if err := c.habitService.RemoveFromGroup(ctx.Context(), hab, group); err != nil {
		c.logger.Error("failed to remove habit from group", zap.Error(err))
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	return ctx.SendStatus(http.StatusOK)
}

func (c Controller) middlewareFindHabit(ctx *fiber.Ctx) error {
	id := ctx.Params("habitID")
	if !IsUUID(id) {
		return ctx.SendStatus(http.StatusNotFound)
	}

	userID := ctx.Locals("userID").(string)

	h, err := c.habitService.Find(ctx.Context(), FindHabitDTO{HabitID: id, UserID: userID})
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

func (c Controller) middlewareFindGroup(ctx *fiber.Ctx) error {
	userID := ctx.Locals("userID").(string)
	groupID := ctx.Params("groupID")
	if !IsUUID(groupID) {
		return ctx.SendStatus(http.StatusNotFound)
	}

	group, err := c.habitService.FindGroup(ctx.Context(), FindGroupDTO{UserID: userID, GroupID: groupID})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ctx.SendStatus(http.StatusNotFound)
		}

		c.logger.Error("failed to get group", zap.Error(err))
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	ctx.Locals("group", group)

	return ctx.Next()
}

func (c Controller) middlewareFindActivity(ctx *fiber.Ctx) error {
	userID := ctx.Locals("userID").(string)
	habitID := ctx.Params("habitID")
	activityID := ctx.Params("activityID")

	if !IsUUID(habitID) || !IsUUID(activityID) {
		return ctx.SendStatus(http.StatusNotFound)
	}

	find := FindActivityDTO{HabitID: habitID, ActivityID: activityID, UserID: userID}
	activity, err := c.habitService.FindActivity(ctx.Context(), find)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ctx.SendStatus(http.StatusNotFound)
		}
		return ctx.SendStatus(http.StatusBadRequest)
	}

	ctx.Locals("activity", activity)

	return ctx.Next()
}

func (c Controller) middlewareDecodeToken(ctx *fiber.Ctx) error {
	tok, ok := ctx.GetReqHeaders()["Authorization"]
	if !ok {
		return ctx.Status(http.StatusUnauthorized).SendString("missing Authorization token")
	}

	id, err := c.tokenService.IDFromToken([]byte(tok))
	if err != nil {
		return ctx.SendStatus(http.StatusUnauthorized)
	}

	ctx.Locals("userID", string(id))
	return ctx.Next()
}

type CheckInPayload struct {
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
}

type NamePayload struct {
	Name string `json:"name"`
}

type GroupsAndHabitsPayload struct {
	Groups []Group `json:"groups"`
	Habits []Habit `json:"habits"`
}

type HabitIDPayload struct {
	HabitID string `json:"habitID"`
}
