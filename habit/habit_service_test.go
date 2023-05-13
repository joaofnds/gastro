package habit_test

import (
	"astro/adapters/logger"
	"astro/adapters/postgres"
	"astro/config"
	"astro/habit"
	"astro/test"
	. "astro/test/matchers"
	"astro/test/transaction"

	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

var _ = Describe("habit service", func() {
	var (
		ctx        context.Context
		app        *fxtest.App
		sut        *habit.HabitService
		userID     = "26b67b16-f8e7-4686-8c78-bc7f5a70ed1a"
		badHabitID = "76767d2e-57f8-41c5-b34f-7b845a084d63"
		uuidLen    = 36
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
			fx.Populate(&sut),
		)
		app.RequireStart()
	})

	AfterEach(func() {
		app.RequireStop()
	})

	Describe("DeleteAll", func() {
		It("removes all habits for the user", func() {
			Must2(sut.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))
			Expect(sut.List(ctx, userID)).NotTo(BeEmpty())

			Must(sut.DeleteAll(ctx))

			Expect(sut.List(ctx, userID)).To(BeEmpty())
		})
	})

	Describe("create", func() {
		Describe("attributes", func() {
			It("Has an ID", func() {
				hab := Must2(sut.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))
				Expect(hab.ID).To(HaveLen(uuidLen))
			})

			It("Has user ID", func() {
				hab := Must2(sut.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))
				Expect(hab.UserID).To(Equal(userID))
			})

			It("Has a name", func() {
				hab := Must2(sut.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))
				Expect(hab.Name).To(Equal("read"))
			})

			It("Has empty activities", func() {
				hab := Must2(sut.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))
				Expect(hab.Activities).To(HaveLen(0))
			})
		})

		It("can be found by id", func() {
			created := Must2(sut.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))
			found := Must2(sut.Find(ctx, habit.FindHabitDTO{HabitID: created.ID, UserID: userID}))

			Expect(found).To(Equal(created))
		})

		It("appear on habits listing", func() {
			hab := Must2(sut.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))

			habitList := Must2(sut.List(ctx, userID))
			Expect(habitList).To(HaveLen(1))
			Expect(habitList[0]).To(Equal(hab))
		})
	})

	Describe("find", func() {
		It("finds the habit", func() {
			created := Must2(sut.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))

			found := Must2(sut.Find(ctx, habit.FindHabitDTO{HabitID: created.ID, UserID: userID}))
			Expect(found).To(Equal(created))
		})

		It("returns NotFoundErr when not found", func() {
			_, err := sut.Find(ctx, habit.FindHabitDTO{HabitID: badHabitID, UserID: userID})
			Expect(err).To(MatchError(habit.ErrNotFound))
		})
	})

	Describe("edit", func() {
		It("changes the name of the habit", func() {
			created := Must2(sut.Create(ctx, habit.CreateHabitDTO{Name: "old", UserID: userID}))
			Must(sut.Update(ctx, habit.UpdateHabitDTO{Name: "new", HabitID: created.ID}))
			found := Must2(sut.Find(ctx, habit.FindHabitDTO{HabitID: created.ID, UserID: userID}))

			Expect(found.Name).To(Equal("new"))
		})

		Describe("with invalid uuid", func() {
			It("returns repo err", func() {
				err := sut.Update(ctx, habit.UpdateHabitDTO{Name: "new", HabitID: "invalid uuid"})

				Expect(err).To(MatchError(habit.ErrRepository))
			})
		})

		Describe("when not found", func() {
			It("returns habit not found err", func() {
				err := sut.Update(ctx, habit.UpdateHabitDTO{Name: "new", HabitID: badHabitID})

				Expect(err).To(MatchError(habit.ErrNotFound))
			})
		})
	})

	Describe("removed habits", func() {
		It("is not present on habit list", func() {
			hab := Must2(sut.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))

			Expect(Must2(sut.List(ctx, userID))).To(HaveLen(1))

			Must(sut.Delete(ctx, habit.FindHabitDTO{HabitID: hab.ID, UserID: userID}))

			Expect(Must2(sut.List(ctx, userID))).To(HaveLen(0))
		})

		It("cannot remove habits for other users", func() {
			otherUserID := "6e334403-1eac-4cf3-bbaf-a9ef4486477a"

			hab := Must2(sut.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))
			err := sut.Delete(ctx, habit.FindHabitDTO{HabitID: hab.ID, UserID: otherUserID})

			Expect(err).To(MatchError(habit.ErrNotFound))
		})
	})
})
