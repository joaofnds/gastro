package habit_test

import (
	"astro/config"
	"astro/habit"
	"astro/postgres"
	"astro/test"
	. "astro/test/matchers"
	"astro/test/transaction"
	"time"

	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestHabit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Habit Service Test")
}

var _ = Describe("habit service", func() {
	var (
		app          *fxtest.App
		habitService *habit.HabitService
		userID       = "26b67b16-f8e7-4686-8c78-bc7f5a70ed1a"
		badHabitID   = "76767d2e-57f8-41c5-b34f-7b845a084d63"
		uuidLen      = 36
	)

	BeforeEach(func() {
		app = fxtest.New(
			GinkgoT(),
			test.NopLogger,
			test.NopHabitInstrumentation,
			test.RandomAppConfigPort,
			config.Module,
			postgres.Module,
			habit.Module,
			fx.Populate(&habitService),
			transaction.Module,
		)
		app.RequireStart()
	})

	AfterEach(func() {
		app.RequireStop()
	})

	Describe("DeleteAll", func() {
		It("removes all habits for the user", func() {
			ctx := context.Background()
			Must2(habitService.Create(ctx, habit.CreateDTO{"read", userID}))
			Expect(habitService.List(ctx, userID)).NotTo(BeEmpty())

			Must(habitService.DeleteAll(ctx))

			Expect(habitService.List(ctx, userID)).To(BeEmpty())
		})
	})

	Describe("create", func() {
		Describe("attributes", func() {
			It("Has an ID", func() {
				ctx := context.Background()
				habit := Must2(habitService.Create(ctx, habit.CreateDTO{"read", userID}))
				Expect(habit.ID).To(HaveLen(uuidLen))
			})

			It("Has user ID", func() {
				ctx := context.Background()
				habit := Must2(habitService.Create(ctx, habit.CreateDTO{"read", userID}))
				Expect(habit.UserID).To(Equal(userID))
			})

			It("Has a name", func() {
				ctx := context.Background()
				habit := Must2(habitService.Create(ctx, habit.CreateDTO{"read", userID}))
				Expect(habit.Name).To(Equal("read"))
			})

			It("Has empty activities", func() {
				ctx := context.Background()
				habit := Must2(habitService.Create(ctx, habit.CreateDTO{"read", userID}))
				Expect(habit.Activities).To(HaveLen(0))
			})
		})

		It("can be found by id", func() {
			ctx := context.Background()
			habitCreated := Must2(habitService.Create(ctx, habit.CreateDTO{"read", userID}))

			habitFound := Must2(habitService.Find(ctx, habit.FindDTO{HabitID: habitCreated.ID, UserID: userID}))

			Expect(habitFound).To(Equal(habitCreated))
		})

		It("has no activities", func() {
			ctx := context.Background()
			habit := Must2(habitService.Create(ctx, habit.CreateDTO{"read", userID}))

			Expect(habit.Activities).To(HaveLen(0))
		})

		It("appear on habits listing", func() {
			ctx := context.Background()
			habit := Must2(habitService.Create(ctx, habit.CreateDTO{"read", userID}))

			habits := Must2(habitService.List(ctx, userID))
			Expect(habits).To(HaveLen(1))
			Expect(habits[0]).To(Equal(habit))
		})
	})

	Describe("Find", func() {
		It("finds the habit", func() {
			ctx := context.Background()
			habitCreated := Must2(habitService.Create(ctx, habit.CreateDTO{"read", userID}))

			habitFound := Must2(habitService.Find(ctx, habit.FindDTO{HabitID: habitCreated.ID, UserID: userID}))
			Expect(habitFound).To(Equal(habitCreated))
		})

		It("returns HabitNotFoundErr when not found", func() {
			ctx := context.Background()
			_, err := habitService.Find(ctx, habit.FindDTO{HabitID: badHabitID, UserID: userID})
			Expect(err).To(MatchError(habit.HabitNotFoundErr))
		})
	})

	Describe("add activity", func() {
		It("persists the activity to the habit", func() {
			ctx := context.Background()
			h := Must2(habitService.Create(ctx, habit.CreateDTO{"read", userID}))

			h = Must2(habitService.Find(ctx, habit.FindDTO{HabitID: h.ID, UserID: userID}))
			Expect(h.Activities).To(HaveLen(0))

			Must2(habitService.AddActivity(ctx, h, time.Now()))

			h = Must2(habitService.Find(ctx, habit.FindDTO{HabitID: h.ID, UserID: userID}))
			Expect(h.Activities).To(HaveLen(1))
		})

		It("sets the provided timestamp truncated to the second", func() {
			ctx := context.Background()
			h := Must2(habitService.Create(ctx, habit.CreateDTO{"read", userID}))

			date := time.Now().UTC()
			Must2(habitService.AddActivity(ctx, h, date))

			h = Must2(habitService.Find(ctx, habit.FindDTO{HabitID: h.ID, UserID: userID}))
			Expect(h.Activities[0].CreatedAt.UTC()).To(Equal(date.Truncate(time.Second)))
		})
	})

	It("removed habits do not appear on habits listing", func() {
		ctx := context.Background()
		h := Must2(habitService.Create(ctx, habit.CreateDTO{"read", userID}))

		Expect(Must2(habitService.List(ctx, userID))).To(HaveLen(1))

		Must(habitService.Delete(ctx, habit.FindDTO{HabitID: h.ID, UserID: userID}))

		Expect(Must2(habitService.List(ctx, userID))).To(HaveLen(0))
	})

	It("cannot remove habits for other users", func() {
		ctx := context.Background()
		otherUserID := "6e334403-1eac-4cf3-bbaf-a9ef4486477a"

		h := Must2(habitService.Create(ctx, habit.CreateDTO{"read", userID}))
		err := habitService.Delete(ctx, habit.FindDTO{HabitID: h.ID, UserID: otherUserID})

		Expect(err).To(MatchError(habit.HabitNotFoundErr))
	})
})
