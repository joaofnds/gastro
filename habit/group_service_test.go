package habit_test

import (
	"astro/adapters/logger"
	"astro/adapters/sqlite"
	"astro/config"
	"astro/habit"
	"astro/test"
	"astro/test/database"
	. "astro/test/matchers"

	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"gorm.io/gorm"
)

var _ = Describe("group service", func() {
	var (
		ctx          context.Context
		app          *fxtest.App
		habitService *habit.HabitService
		sut          *habit.GroupService
		orm          *gorm.DB
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
			sqlite.Module,
			habit.Module,
			database.Module,
			fx.Populate(&sut, &habitService, &orm),
		)
		app.RequireStart()
	})

	AfterEach(func() {
		app.RequireStop()
	})

	Describe("created groups", func() {
		It("deletes only the owning user's group when an imported ID is shared", func() {
			group := habit.Group{ID: "12767d2e-57f8-41c5-b34f-7b845a084d63", UserID: userID, Name: "health", Habits: []habit.Habit{}}
			other := group
			other.UserID = "76b67b16-f8e7-4686-8c78-bc7f5a70ed1a"
			Expect(orm.Create(&group).Error).To(Succeed())
			Expect(orm.Create(&other).Error).To(Succeed())

			Expect(sut.Delete(ctx, group)).To(Succeed())

			_, err := sut.Find(ctx, habit.FindGroupDTO{GroupID: group.ID, UserID: group.UserID})
			Expect(err).To(MatchError(habit.ErrNotFound))
			found, err := sut.Find(ctx, habit.FindGroupDTO{GroupID: other.ID, UserID: other.UserID})
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(Equal(other))
		})

		It("are listed", func() {
			health := Must2(sut.Create(ctx, habit.CreateGroupDTO{UserID: userID, Name: "health"}))

			groups, _ := Must3(sut.GroupsAndHabits(ctx, userID))

			Expect(groups).To(Equal([]habit.Group{health}))
		})

		It("can be empty", func() {
			health := Must2(sut.Create(ctx, habit.CreateGroupDTO{UserID: userID, Name: "health"}))

			groups, _ := Must3(sut.GroupsAndHabits(ctx, userID))

			Expect(groups).To(Equal([]habit.Group{health}))
			Expect(groups[0].Habits).To(BeEmpty())
		})

		It("do not include habits not associated with them", func() {
			health := Must2(sut.Create(ctx, habit.CreateGroupDTO{UserID: userID, Name: "health"}))
			read := Must2(habitService.Create(ctx, habit.CreateHabitDTO{UserID: userID, Name: "read"}))

			groups, habits := Must3(sut.GroupsAndHabits(ctx, userID))

			Expect(groups).To(Equal([]habit.Group{health}))
			Expect(groups[0].Habits).To(BeEmpty())
			Expect(habits).To(Equal([]habit.Habit{read}))
		})

		It("can be found by ID", func() {
			created := Must2(sut.Create(ctx, habit.CreateGroupDTO{UserID: userID, Name: "health"}))
			found := Must2(sut.Find(ctx, habit.FindGroupDTO{UserID: userID, GroupID: created.ID}))

			Expect(found).To(Equal(created))
		})
	})

	Describe("habit in a group", func() {
		It("is listed in the group", func() {
			health := Must2(sut.Create(ctx, habit.CreateGroupDTO{UserID: userID, Name: "health"}))
			run := Must2(habitService.Create(ctx, habit.CreateHabitDTO{UserID: userID, Name: "run"}))

			Must(sut.Join(ctx, run, health))

			groups, _ := Must3(sut.GroupsAndHabits(ctx, userID))

			Expect(groups).To(HaveLen(1))
			Expect(groups[0].Habits).To(Equal([]habit.Habit{run}))
		})

		It("is not listed in habits", func() {
			health := Must2(sut.Create(ctx, habit.CreateGroupDTO{UserID: userID, Name: "health"}))
			run := Must2(habitService.Create(ctx, habit.CreateHabitDTO{UserID: userID, Name: "run"}))

			Must(sut.Join(ctx, run, health))

			_, habits := Must3(sut.GroupsAndHabits(ctx, userID))

			Expect(habits).NotTo(ContainElements(run))
		})

		When("removed from group", func() {
			It("moves from 'groups' to 'habits'", func() {
				health := Must2(sut.Create(ctx, habit.CreateGroupDTO{UserID: userID, Name: "health"}))
				run := Must2(habitService.Create(ctx, habit.CreateHabitDTO{UserID: userID, Name: "run"}))

				Must(sut.Join(ctx, run, health))
				groups, habits := Must3(sut.GroupsAndHabits(ctx, userID))
				Expect(groups[0].Habits).To(ContainElement(run))
				Expect(habits).NotTo(ContainElements(run))

				Must(sut.Leave(ctx, run, health))
				groups, habits = Must3(sut.GroupsAndHabits(ctx, userID))
				Expect(groups[0].Habits).NotTo(ContainElement(run))
				Expect(habits).To(ContainElements(run))
			})
		})

		When("group is deleted", func() {
			It("keeps habits", func() {
				health := Must2(sut.Create(ctx, habit.CreateGroupDTO{UserID: userID, Name: "health"}))
				run := Must2(habitService.Create(ctx, habit.CreateHabitDTO{UserID: userID, Name: "run"}))
				gym := Must2(habitService.Create(ctx, habit.CreateHabitDTO{UserID: userID, Name: "gym"}))
				read := Must2(habitService.Create(ctx, habit.CreateHabitDTO{UserID: userID, Name: "read"}))

				Must(sut.Join(ctx, run, health))
				Must(sut.Join(ctx, gym, health))

				Must(sut.Delete(ctx, health))

				groups, habits := Must3(sut.GroupsAndHabits(ctx, userID))

				Expect(groups).To(BeEmpty())
				Expect(habits).To(ContainElements(read, run, gym))
			})
		})
	})

	It("groups and habits live happy together", func() {
		health := Must2(sut.Create(ctx, habit.CreateGroupDTO{UserID: userID, Name: "health"}))
		run := Must2(habitService.Create(ctx, habit.CreateHabitDTO{UserID: userID, Name: "run"}))
		gym := Must2(habitService.Create(ctx, habit.CreateHabitDTO{UserID: userID, Name: "gym"}))
		read := Must2(habitService.Create(ctx, habit.CreateHabitDTO{UserID: userID, Name: "read"}))

		Must(sut.Join(ctx, run, health))
		Must(sut.Join(ctx, gym, health))

		groups, habits := Must3(sut.GroupsAndHabits(ctx, userID))

		Expect(habits).To(Equal([]habit.Habit{read}))

		Expect(groups).To(HaveLen(1))
		Expect(groups[0].Habits).To(ContainElements(run, gym))
	})
})
