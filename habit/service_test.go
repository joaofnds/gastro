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
		habitService *habit.Service
		userID       = "26b67b16-f8e7-4686-8c78-bc7f5a70ed1a"
		badHabitID   = "76767d2e-57f8-41c5-b34f-7b845a084d63"
		uuidLen      = 36
	)

	BeforeEach(func() {
		app = fxtest.New(
			GinkgoT(),
			test.NopLogger,
			test.NopHabitInstrumentation,
			test.NewPortAppConfig,
			config.Module,
			postgres.Module,
			habit.Module,
			transaction.Module,
			fx.Populate(&habitService),
		)
		app.RequireStart()
	})

	AfterEach(func() {
		app.RequireStop()
	})

	Describe("DeleteAll", func() {
		It("removes all habits for the user", func() {
			ctx := context.Background()
			Must2(habitService.Create(ctx, habit.CreateHabitDTO{"read", userID}))
			Expect(habitService.List(ctx, userID)).NotTo(BeEmpty())

			Must(habitService.DeleteAll(ctx))

			Expect(habitService.List(ctx, userID)).To(BeEmpty())
		})
	})

	Describe("create", func() {
		Describe("attributes", func() {
			It("Has an ID", func() {
				ctx := context.Background()
				habit := Must2(habitService.Create(ctx, habit.CreateHabitDTO{"read", userID}))
				Expect(habit.ID).To(HaveLen(uuidLen))
			})

			It("Has user ID", func() {
				ctx := context.Background()
				habit := Must2(habitService.Create(ctx, habit.CreateHabitDTO{"read", userID}))
				Expect(habit.UserID).To(Equal(userID))
			})

			It("Has a name", func() {
				ctx := context.Background()
				habit := Must2(habitService.Create(ctx, habit.CreateHabitDTO{"read", userID}))
				Expect(habit.Name).To(Equal("read"))
			})

			It("Has empty activities", func() {
				ctx := context.Background()
				habit := Must2(habitService.Create(ctx, habit.CreateHabitDTO{"read", userID}))
				Expect(habit.Activities).To(HaveLen(0))
			})
		})

		It("can be found by id", func() {
			ctx := context.Background()
			habitCreated := Must2(habitService.Create(ctx, habit.CreateHabitDTO{"read", userID}))

			habitFound := Must2(habitService.Find(ctx, habit.FindHabitDTO{HabitID: habitCreated.ID, UserID: userID}))

			Expect(habitFound).To(Equal(habitCreated))
		})

		It("appear on habits listing", func() {
			ctx := context.Background()
			habit := Must2(habitService.Create(ctx, habit.CreateHabitDTO{"read", userID}))

			habits := Must2(habitService.List(ctx, userID))
			Expect(habits).To(HaveLen(1))
			Expect(habits[0]).To(Equal(habit))
		})
	})

	Describe("Find", func() {
		It("finds the habit", func() {
			ctx := context.Background()
			habitCreated := Must2(habitService.Create(ctx, habit.CreateHabitDTO{"read", userID}))

			habitFound := Must2(habitService.Find(ctx, habit.FindHabitDTO{HabitID: habitCreated.ID, UserID: userID}))
			Expect(habitFound).To(Equal(habitCreated))
		})

		It("returns NotFoundErr when not found", func() {
			ctx := context.Background()
			_, err := habitService.Find(ctx, habit.FindHabitDTO{HabitID: badHabitID, UserID: userID})
			Expect(err).To(MatchError(habit.ErrNotFound))
		})
	})

	Describe("add activity", func() {
		It("creates an uuid", func() {
			ctx := context.Background()
			hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{"read", userID}))
			act := Must2(habitService.AddActivity(ctx, hab, habit.AddActivityDTO{Time: time.Now()}))
			Expect(habit.IsUUID(act.ID)).To(BeTrue())
		})

		It("persists the activity to the habit", func() {
			ctx := context.Background()
			hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{"read", userID}))

			hab = Must2(habitService.Find(ctx, habit.FindHabitDTO{HabitID: hab.ID, UserID: userID}))
			Expect(hab.Activities).To(HaveLen(0))

			Must2(habitService.AddActivity(ctx, hab, habit.AddActivityDTO{Time: time.Now()}))

			hab = Must2(habitService.Find(ctx, habit.FindHabitDTO{HabitID: hab.ID, UserID: userID}))
			Expect(hab.Activities).To(HaveLen(1))
		})

		It("sets the provided timestamp in UTC truncated to the second", func() {
			ctx := context.Background()
			hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{"read", userID}))

			date := time.Now()
			act := Must2(habitService.AddActivity(ctx, hab, habit.AddActivityDTO{Time: date}))

			found := Must2(habitService.FindActivity(ctx, habit.FindActivityDTO{hab.ID, act.ID, userID}))
			Expect(found.CreatedAt).To(Equal(date.UTC().Truncate(time.Second)))
		})

		It("sets the description", func() {
			ctx := context.Background()
			hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{"read", userID}))
			dto := habit.AddActivityDTO{Desc: "my description", Time: time.Now()}
			act := Must2(habitService.AddActivity(ctx, hab, dto))

			found := Must2(habitService.FindActivity(ctx, habit.FindActivityDTO{hab.ID, act.ID, userID}))

			Expect(found.Desc).To(Equal(dto.Desc))
		})
	})

	Describe("find activity", func() {
		It("finds the activity", func() {
			ctx := context.Background()
			hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{"read", userID}))
			act := Must2(habitService.AddActivity(ctx, hab, habit.AddActivityDTO{Desc: "foo", Time: time.Now()}))

			Must2(habitService.FindActivity(ctx, habit.FindActivityDTO{hab.ID, act.ID, userID}))
		})
	})

	Describe("update activity", func() {
		It("updates activity description", func() {
			ctx := context.Background()
			hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{"read", userID}))
			act := Must2(habitService.AddActivity(ctx, hab, habit.AddActivityDTO{Desc: "old", Time: time.Now()}))
			Must2(habitService.UpdateActivity(ctx, habit.UpdateActivityDTO{ActivityID: act.ID, Desc: "new"}))

			act = Must2(habitService.FindActivity(ctx, habit.FindActivityDTO{hab.ID, act.ID, userID}))
			Expect(act.Desc).To(Equal("new"))
		})
	})

	Describe("delete activity", func() {
		It("deletes the activity", func() {
			ctx := context.Background()
			hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{"read", userID}))
			act := Must2(habitService.AddActivity(ctx, hab, habit.AddActivityDTO{Time: time.Now()}))
			hab = Must2(habitService.Find(ctx, habit.FindHabitDTO{hab.ID, userID}))
			Expect(hab.Activities).To(HaveLen(1))

			Must(habitService.DeleteActivity(ctx, act))

			hab = Must2(habitService.Find(ctx, habit.FindHabitDTO{hab.ID, userID}))
			Expect(hab.Activities).To(HaveLen(0))
		})
	})

	It("removed habits do not appear on habits listing", func() {
		ctx := context.Background()
		h := Must2(habitService.Create(ctx, habit.CreateHabitDTO{"read", userID}))

		Expect(Must2(habitService.List(ctx, userID))).To(HaveLen(1))

		Must(habitService.Delete(ctx, habit.FindHabitDTO{HabitID: h.ID, UserID: userID}))

		Expect(Must2(habitService.List(ctx, userID))).To(HaveLen(0))
	})

	It("cannot remove habits for other users", func() {
		ctx := context.Background()
		otherUserID := "6e334403-1eac-4cf3-bbaf-a9ef4486477a"

		h := Must2(habitService.Create(ctx, habit.CreateHabitDTO{"read", userID}))
		err := habitService.Delete(ctx, habit.FindHabitDTO{HabitID: h.ID, UserID: otherUserID})

		Expect(err).To(MatchError(habit.ErrNotFound))
	})
})
