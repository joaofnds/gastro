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

var _ = Describe("group service", func() {
	var (
		ctx          context.Context
		app          *fxtest.App
		habitService *habit.HabitService
		sut          *habit.GroupService
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

	Describe("created groups", func() {
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
