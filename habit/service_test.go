package habit_test

import (
	"astro/config"
	"astro/habit"
	"astro/logger"
	"astro/postgres"
	"astro/test"
	. "astro/test/matchers"
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
	var app *fxtest.App
	var habitService *habit.HabitService

	BeforeEach(func() {
		app = fxtest.New(
			GinkgoT(),
			fx.NopLogger,
			logger.Module,
			config.Module,
			fx.Decorate(test.RandomAppConfigPort),
			postgres.Module,
			habit.Module,
			fx.Populate(&habitService),
		)
		app.RequireStart()
		Must(habitService.DeleteAll(context.Background()))
	})

	AfterEach(func() {
		app.RequireStop()
	})

	Describe("DeleteAll", func() {
		It("removes all habits", func() {
			ctx := context.Background()
			Must2(habitService.Create(ctx, "read"))
			Expect(habitService.List(ctx)).NotTo(BeEmpty())

			Must(habitService.DeleteAll(ctx))

			Expect(habitService.List(ctx)).To(BeEmpty())
		})
	})

	Describe("create", func() {
		It("Has an ID", func() {
			ctx := context.Background()
			habit := Must2(habitService.Create(ctx, "read"))

			Expect(habit.ID > 0).To(BeTrue())
		})

		It("can be found by name", func() {
			ctx := context.Background()
			habit := Must2(habitService.Create(ctx, "read"))

			habitFound := Must2(habitService.FindByName(ctx, habit.Name))

			Expect(habitFound).To(Equal(habit))
		})

		It("appear on habits listing", func() {
			ctx := context.Background()
			habit := Must2(habitService.Create(ctx, "read"))

			habits := Must2(habitService.List(ctx))
			Expect(habits).To(HaveLen(1))
			Expect(habits[0]).To(Equal(habit))
		})
	})

	Describe("add activity", func() {
		It("persists the activity to the habit", func() {
			ctx := context.Background()
			Must2(habitService.Create(ctx, "read"))

			habit := Must2(habitService.FindByName(ctx, "read"))
			Expect(habit.Activities).To(HaveLen(0))

			Must2(habitService.AddActivity(ctx, habit, time.Now()))

			habit = Must2(habitService.FindByName(ctx, habit.Name))
			Expect(habit.Activities).To(HaveLen(1))
		})

		It("sets the provided timestamp truncated to the second", func() {
			ctx := context.Background()
			habit := Must2(habitService.Create(ctx, "read"))

			date := time.Now().UTC()
			Must2(habitService.AddActivity(ctx, habit, date))

			habit = Must2(habitService.FindByName(ctx, habit.Name))
			Expect(habit.Activities[0].CreatedAt.UTC()).To(Equal(date.Truncate(time.Second)))
		})
	})

	It("removed habits do not appear on habits listing", func() {
		ctx := context.Background()
		habit := Must2(habitService.Create(ctx, "read"))

		Expect(Must2(habitService.List(ctx))).To(HaveLen(1))

		Must(habitService.DeleteByName(ctx, habit.Name))

		Expect(Must2(habitService.List(ctx))).To(HaveLen(0))
	})
})
