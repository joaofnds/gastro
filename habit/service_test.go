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
		ctx          context.Context
		app          *fxtest.App
		habitService *habit.Service
		userID       = "26b67b16-f8e7-4686-8c78-bc7f5a70ed1a"
		badHabitID   = "76767d2e-57f8-41c5-b34f-7b845a084d63"
		uuidLen      = 36
	)

	BeforeEach(func() {
		ctx = context.Background()
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
			Must2(habitService.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))
			Expect(habitService.List(ctx, userID)).NotTo(BeEmpty())

			Must(habitService.DeleteAll(ctx))

			Expect(habitService.List(ctx, userID)).To(BeEmpty())
		})
	})

	Describe("create", func() {
		Describe("attributes", func() {
			It("Has an ID", func() {
				hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))
				Expect(hab.ID).To(HaveLen(uuidLen))
			})

			It("Has user ID", func() {
				hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))
				Expect(hab.UserID).To(Equal(userID))
			})

			It("Has a name", func() {
				hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))
				Expect(hab.Name).To(Equal("read"))
			})

			It("Has empty activities", func() {
				hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))
				Expect(hab.Activities).To(HaveLen(0))
			})
		})

		It("can be found by id", func() {
			created := Must2(habitService.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))
			found := Must2(habitService.Find(ctx, habit.FindHabitDTO{HabitID: created.ID, UserID: userID}))

			Expect(found).To(Equal(created))
		})

		It("appear on habits listing", func() {
			hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))

			habitList := Must2(habitService.List(ctx, userID))
			Expect(habitList).To(HaveLen(1))
			Expect(habitList[0]).To(Equal(hab))
		})
	})

	Describe("Find", func() {
		It("finds the habit", func() {
			created := Must2(habitService.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))

			found := Must2(habitService.Find(ctx, habit.FindHabitDTO{HabitID: created.ID, UserID: userID}))
			Expect(found).To(Equal(created))
		})

		It("returns NotFoundErr when not found", func() {
			_, err := habitService.Find(ctx, habit.FindHabitDTO{HabitID: badHabitID, UserID: userID})
			Expect(err).To(MatchError(habit.ErrNotFound))
		})
	})

	Describe("edit", func() {
		It("changes the name of the habit", func() {
			created := Must2(habitService.Create(ctx, habit.CreateHabitDTO{Name: "old", UserID: userID}))
			Must(habitService.Update(ctx, habit.UpdateHabitDTO{Name: "new", HabitID: created.ID}))
			found := Must2(habitService.Find(ctx, habit.FindHabitDTO{HabitID: created.ID, UserID: userID}))

			Expect(found.Name).To(Equal("new"))
		})

		Describe("with invalid uuid", func() {
			It("returns repo err", func() {
				err := habitService.Update(ctx, habit.UpdateHabitDTO{Name: "new", HabitID: "invalid uuid"})

				Expect(err).To(MatchError(habit.ErrRepository))
			})
		})

		Describe("when not found", func() {
			It("returns habit not found err", func() {
				err := habitService.Update(ctx, habit.UpdateHabitDTO{Name: "new", HabitID: badHabitID})

				Expect(err).To(MatchError(habit.ErrNotFound))
			})
		})
	})

	Describe("add activity", func() {
		It("creates an uuid", func() {
			hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))
			act := Must2(habitService.AddActivity(ctx, hab, habit.AddActivityDTO{Time: time.Now()}))
			Expect(habit.IsUUID(act.ID)).To(BeTrue())
		})

		It("persists the activity to the habit", func() {
			hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))

			hab = Must2(habitService.Find(ctx, habit.FindHabitDTO{HabitID: hab.ID, UserID: userID}))
			Expect(hab.Activities).To(HaveLen(0))

			Must2(habitService.AddActivity(ctx, hab, habit.AddActivityDTO{Time: time.Now()}))

			hab = Must2(habitService.Find(ctx, habit.FindHabitDTO{HabitID: hab.ID, UserID: userID}))
			Expect(hab.Activities).To(HaveLen(1))
		})

		It("sets the provided timestamp in UTC truncated to the second", func() {
			hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))

			date := time.Now()
			act := Must2(habitService.AddActivity(ctx, hab, habit.AddActivityDTO{Time: date}))

			found := Must2(habitService.FindActivity(ctx, habit.FindActivityDTO{HabitID: hab.ID, ActivityID: act.ID, UserID: userID}))
			Expect(found.CreatedAt).To(Equal(date.UTC().Truncate(time.Second)))
		})

		It("sets the description", func() {
			hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))
			dto := habit.AddActivityDTO{Desc: "my description", Time: time.Now()}
			act := Must2(habitService.AddActivity(ctx, hab, dto))

			found := Must2(habitService.FindActivity(ctx, habit.FindActivityDTO{HabitID: hab.ID, ActivityID: act.ID, UserID: userID}))

			Expect(found.Desc).To(Equal(dto.Desc))
		})
	})

	Describe("find activity", func() {
		It("finds the activity", func() {
			hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))
			act := Must2(habitService.AddActivity(ctx, hab, habit.AddActivityDTO{Desc: "foo", Time: time.Now()}))

			Must2(habitService.FindActivity(ctx, habit.FindActivityDTO{HabitID: hab.ID, ActivityID: act.ID, UserID: userID}))
		})
	})

	Describe("update activity", func() {
		It("updates activity description", func() {
			hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))
			act := Must2(habitService.AddActivity(ctx, hab, habit.AddActivityDTO{Desc: "old", Time: time.Now()}))
			Must2(habitService.UpdateActivity(ctx, habit.UpdateActivityDTO{ActivityID: act.ID, Desc: "new"}))

			act = Must2(habitService.FindActivity(ctx, habit.FindActivityDTO{HabitID: hab.ID, ActivityID: act.ID, UserID: userID}))
			Expect(act.Desc).To(Equal("new"))
		})
	})

	Describe("delete activity", func() {
		It("deletes the activity", func() {
			hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))
			act := Must2(habitService.AddActivity(ctx, hab, habit.AddActivityDTO{Time: time.Now()}))
			hab = Must2(habitService.Find(ctx, habit.FindHabitDTO{HabitID: hab.ID, UserID: userID}))
			Expect(hab.Activities).To(HaveLen(1))

			Must(habitService.DeleteActivity(ctx, act))

			hab = Must2(habitService.Find(ctx, habit.FindHabitDTO{HabitID: hab.ID, UserID: userID}))
			Expect(hab.Activities).To(HaveLen(0))
		})
	})

	Describe("removed habits", func() {
		It("is not present on habit list", func() {
			hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))

			Expect(Must2(habitService.List(ctx, userID))).To(HaveLen(1))

			Must(habitService.Delete(ctx, habit.FindHabitDTO{HabitID: hab.ID, UserID: userID}))

			Expect(Must2(habitService.List(ctx, userID))).To(HaveLen(0))
		})

		It("cannot remove habits for other users", func() {
			otherUserID := "6e334403-1eac-4cf3-bbaf-a9ef4486477a"

			hab := Must2(habitService.Create(ctx, habit.CreateHabitDTO{Name: "read", UserID: userID}))
			err := habitService.Delete(ctx, habit.FindHabitDTO{HabitID: hab.ID, UserID: otherUserID})

			Expect(err).To(MatchError(habit.ErrNotFound))
		})
	})

	Describe("groups", func() {
		Describe("created groups", func() {
			It("are listed", func() {
				health := Must2(habitService.CreateGroup(ctx, habit.CreateGroupDTO{UserID: userID, Name: "health"}))

				groups, _ := Must3(habitService.GroupsAndHabits(ctx, userID))

				Expect(groups).To(Equal([]habit.Group{health}))
			})

			It("can be empty", func() {
				health := Must2(habitService.CreateGroup(ctx, habit.CreateGroupDTO{UserID: userID, Name: "health"}))

				groups, _ := Must3(habitService.GroupsAndHabits(ctx, userID))

				Expect(groups).To(Equal([]habit.Group{health}))
				Expect(groups[0].Habits).To(BeEmpty())
			})

			It("do not include habits not associated with them", func() {
				health := Must2(habitService.CreateGroup(ctx, habit.CreateGroupDTO{UserID: userID, Name: "health"}))
				read := Must2(habitService.Create(ctx, habit.CreateHabitDTO{UserID: userID, Name: "read"}))

				groups, habits := Must3(habitService.GroupsAndHabits(ctx, userID))

				Expect(groups).To(Equal([]habit.Group{health}))
				Expect(groups[0].Habits).To(BeEmpty())
				Expect(habits).To(Equal([]habit.Habit{read}))
			})

			It("can be found by ID", func() {
				created := Must2(habitService.CreateGroup(ctx, habit.CreateGroupDTO{UserID: userID, Name: "health"}))
				found := Must2(habitService.FindGroup(ctx, habit.FindGroupDTO{UserID: userID, GroupID: created.ID}))

				Expect(found).To(Equal(created))
			})
		})

		Describe("group habit", func() {
			It("is listed in the group", func() {
				health := Must2(habitService.CreateGroup(ctx, habit.CreateGroupDTO{UserID: userID, Name: "health"}))
				run := Must2(habitService.Create(ctx, habit.CreateHabitDTO{UserID: userID, Name: "run"}))

				Must(habitService.AddToGroup(ctx, run, health))

				groups, _ := Must3(habitService.GroupsAndHabits(ctx, userID))

				Expect(groups).To(HaveLen(1))
				Expect(groups[0].Habits).To(Equal([]habit.Habit{run}))
			})

			It("is not listed in habits", func() {
				health := Must2(habitService.CreateGroup(ctx, habit.CreateGroupDTO{UserID: userID, Name: "health"}))
				run := Must2(habitService.Create(ctx, habit.CreateHabitDTO{UserID: userID, Name: "run"}))

				Must(habitService.AddToGroup(ctx, run, health))

				_, habits := Must3(habitService.GroupsAndHabits(ctx, userID))

				Expect(habits).NotTo(ContainElements(run))
			})

			When("removed from group", func() {
				It("moves from 'groups' to 'habits'", func() {
					health := Must2(habitService.CreateGroup(ctx, habit.CreateGroupDTO{UserID: userID, Name: "health"}))
					run := Must2(habitService.Create(ctx, habit.CreateHabitDTO{UserID: userID, Name: "run"}))

					Must(habitService.AddToGroup(ctx, run, health))
					groups, habits := Must3(habitService.GroupsAndHabits(ctx, userID))
					Expect(groups[0].Habits).To(ContainElement(run))
					Expect(habits).NotTo(ContainElements(run))

					Must(habitService.RemoveFromGroup(ctx, run, health))
					groups, habits = Must3(habitService.GroupsAndHabits(ctx, userID))
					Expect(groups[0].Habits).NotTo(ContainElement(run))
					Expect(habits).To(ContainElements(run))
				})
			})
		})

		It("groups and habits live happy together", func() {
			health := Must2(habitService.CreateGroup(ctx, habit.CreateGroupDTO{UserID: userID, Name: "health"}))
			run := Must2(habitService.Create(ctx, habit.CreateHabitDTO{UserID: userID, Name: "run"}))
			cycle := Must2(habitService.Create(ctx, habit.CreateHabitDTO{UserID: userID, Name: "cycle"}))
			read := Must2(habitService.Create(ctx, habit.CreateHabitDTO{UserID: userID, Name: "read"}))

			Must(habitService.AddToGroup(ctx, run, health))
			Must(habitService.AddToGroup(ctx, cycle, health))

			groups, habits := Must3(habitService.GroupsAndHabits(ctx, userID))

			Expect(habits).To(Equal([]habit.Habit{read}))

			Expect(groups).To(HaveLen(1))
			Expect(groups[0].Habits).To(ContainElements(run, cycle))
		})
	})
})
