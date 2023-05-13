package habit_test

import (
	"astro/adapters/logger"
	"astro/adapters/postgres"
	"astro/config"
	"astro/habit"
	"astro/test"
	. "astro/test/matchers"
	"astro/test/transaction"
	"time"

	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

var _ = Describe("activity service", func() {
	var (
		ctx          context.Context
		app          *fxtest.App
		habitService *habit.HabitService
		sut          *habit.ActivityService
		userID       = "26b67b16-f8e7-4686-8c78-bc7f5a70ed1a"
	)

	BeforeEach(func() {
		ctx = context.Background()
		app = fxtest.New(
			GinkgoT(),
			logger.NopLogger,
			habit.NopProbeProvider,
			test.NewPortAppConfig,
			config.Module,
			postgres.Module,
			habit.Module,
			transaction.Module,
			fx.Populate(&sut, &habitService),
		)
		app.RequireStart()
	})

	AfterEach(func() {
		app.RequireStop()
	})

	Describe("add", func() {
		It("creates an uuid", func() {
			hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))
			act := Must2(sut.Add(ctx, hab, habit.AddActivityDTO{Time: time.Now()}))
			Expect(habit.IsUUID(act.ID)).To(BeTrue())
		})

		It("persists the activity to the habit", func() {
			hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))

			hab = Must2(habitService.Find(ctx, habit.FindHabitDTO{HabitID: hab.ID, UserID: userID}))
			Expect(hab.Activities).To(HaveLen(0))

			Must2(sut.Add(ctx, hab, habit.AddActivityDTO{Time: time.Now()}))

			hab = Must2(habitService.Find(ctx, habit.FindHabitDTO{HabitID: hab.ID, UserID: userID}))
			Expect(hab.Activities).To(HaveLen(1))
		})

		It("sets the provided timestamp in UTC truncated to the second", func() {
			hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))

			date := time.Now()
			act := Must2(sut.Add(ctx, hab, habit.AddActivityDTO{Time: date}))

			found := Must2(sut.Find(ctx, habit.FindActivityDTO{HabitID: hab.ID, ActivityID: act.ID, UserID: userID}))
			Expect(found.CreatedAt).To(Equal(date.UTC().Truncate(time.Second)))
		})

		It("sets the description", func() {
			hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))
			dto := habit.AddActivityDTO{Desc: "my description", Time: time.Now()}
			act := Must2(sut.Add(ctx, hab, dto))

			found := Must2(sut.Find(ctx, habit.FindActivityDTO{HabitID: hab.ID, ActivityID: act.ID, UserID: userID}))

			Expect(found.Desc).To(Equal(dto.Desc))
		})
	})

	Describe("find", func() {
		It("finds the activity", func() {
			hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))
			act := Must2(sut.Add(ctx, hab, habit.AddActivityDTO{Desc: "foo", Time: time.Now()}))

			Must2(sut.Find(ctx, habit.FindActivityDTO{HabitID: hab.ID, ActivityID: act.ID, UserID: userID}))
		})
	})

	Describe("update", func() {
		It("updates activity description", func() {
			hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))
			act := Must2(sut.Add(ctx, hab, habit.AddActivityDTO{Desc: "old", Time: time.Now()}))
			Must2(sut.Update(ctx, habit.UpdateActivityDTO{ActivityID: act.ID, Desc: "new"}))

			act = Must2(sut.Find(ctx, habit.FindActivityDTO{HabitID: hab.ID, ActivityID: act.ID, UserID: userID}))
			Expect(act.Desc).To(Equal("new"))
		})
	})

	Describe("delete", func() {
		It("deletes the activity", func() {
			hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))
			act := Must2(sut.Add(ctx, hab, habit.AddActivityDTO{Time: time.Now()}))
			hab = Must2(habitService.Find(ctx, habit.FindHabitDTO{HabitID: hab.ID, UserID: userID}))
			Expect(hab.Activities).To(HaveLen(1))

			Must(sut.Delete(ctx, act))

			hab = Must2(habitService.Find(ctx, habit.FindHabitDTO{HabitID: hab.ID, UserID: userID}))
			Expect(hab.Activities).To(HaveLen(0))
		})
	})
})
