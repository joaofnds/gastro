package habit_test

import (
	http2 "astro/adapters/http"
	"astro/adapters/logger"
	"astro/adapters/postgres"
	"astro/config"
	"astro/habit"
	"astro/test"
	"astro/test/driver"
	. "astro/test/matchers"
	"astro/test/transaction"
	"astro/token"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

var _ = Describe("/habits", func() {
	var (
		fxApp         *fxtest.App
		app           *driver.Driver
		api           *driver.API
		uuidLen       = 36
		badHabitID    = "76767d2e-57f8-41c5-b34f-7b845a084d63"
		badActivityID = "76767d2e-57f8-41c5-b34f-7b845a084d64"
	)

	BeforeEach(func() {
		var httpConfig http2.Config
		fxApp = fxtest.New(
			GinkgoT(),
			logger.NopLogger,
			test.NewPortAppConfig,
			habit.NopProbeProvider,
			token.NopProbeProvider,
			http2.NopProbeProvider,
			config.Module,
			http2.FiberModule,
			postgres.Module,
			habit.Module,
			transaction.Module,
			token.Module,
			fx.Invoke(func(app *fiber.App, habitController *habit.Controller, tokenController *token.Controller) {
				habitController.Register(app)
				tokenController.Register(app)
			}),
			fx.Populate(&httpConfig),
		).RequireStart()

		url := fmt.Sprintf("http://localhost:%d", httpConfig.Port)
		app = driver.NewDriver(url)
		api = driver.NewAPI(url)
		app.GetToken()
	})

	AfterEach(func() {
		fxApp.RequireStop()
	})

	Describe("GET", func() {
		It("returns a list of habits", func() {
			app.MustCreate("read")

			data := app.MustList()

			Expect(data).To(HaveLen(1))
			Expect(data[0].Name).To(Equal("read"))
		})

		It("requires token", func() {
			res := api.MustList("")
			Expect(res.StatusCode).To(Equal(http.StatusUnauthorized))
		})
	})

	Describe("POST", func() {
		It("returns status created", func() {
			res := api.MustCreate(app.Token, "read")
			Expect(res.StatusCode).To(Equal(http.StatusCreated))
		})

		It("returns the created habit", func() {
			res := api.MustCreate(app.Token, "read")
			body := Must2(io.ReadAll(res.Body))
			defer func() { _ = res.Body.Close() }()

			var habit habit.Habit
			Must(json.Unmarshal(body, &habit))
			Expect(habit.Name).To(Equal("read"))
		})

		It("requires token", func() {
			res := api.MustCreate("", "read")
			Expect(res.StatusCode).To(Equal(http.StatusUnauthorized))
		})

		Describe("without name", func() {
			It("return bad request", func() {
				res := api.MustCreate(app.Token, "")
				Expect(res.StatusCode).To(Equal(http.StatusBadRequest))
			})
		})
	})

	Describe("/:id", func() {
		It("requires token", func() {
			habit := app.MustCreate("read")

			res := api.MustGet("", habit.ID)
			Expect(res.StatusCode).To(Equal(http.StatusUnauthorized))
		})

		Describe("when habit is found", func() {
			It("has status 200", func() {
				habit := app.MustCreate("read")

				res := api.MustGet(app.Token, habit.ID)
				Expect(res.StatusCode).To(Equal(http.StatusOK))
			})

			It("returns the habit", func() {
				h := app.MustCreate("read")

				habit := app.MustGet(h.ID)
				Expect(habit.ID).To(HaveLen(uuidLen))
				Expect(habit.Name).To(Equal("read"))
				Expect(habit.Activities).To(HaveLen(0))
			})
		})

		Describe("update", func() {
			It("return OK", func() {
				hab := app.MustCreate("read")
				app.MustUpdate(hab.ID, "read")

				res := api.MustGet(app.Token, hab.ID)
				Expect(res.StatusCode).To(Equal(http.StatusOK))
			})

			It("changes the name", func() {
				hab := app.MustCreate("old")
				app.MustUpdate(hab.ID, "new")
				found := app.MustGet(hab.ID)

				Expect(found.Name).To(Equal("new"))
			})

			Describe("with invalid id", func() {
				It("returns not found", func() {
					res := api.MustUpdate(app.Token, "invalid uuid", "name")
					Expect(res.StatusCode).To(Equal(http.StatusNotFound))
				})
			})

			Describe("with habit id that does not exist", func() {
				It("returns not found", func() {
					res := api.MustUpdate(app.Token, badHabitID, "name")
					Expect(res.StatusCode).To(Equal(http.StatusNotFound))
				})
			})
		})

		Describe("after deleting the habit", func() {
			It("has status 404", func() {
				h := app.MustCreate("read")

				res := api.MustGet(app.Token, h.ID)
				Expect(res.StatusCode).To(Equal(http.StatusOK))

				res = api.MustDelete(app.Token, h.ID)
				Expect(res.StatusCode).To(Equal(http.StatusOK))

				res = api.MustGet(app.Token, h.ID)
				Expect(res.StatusCode).To(Equal(http.StatusNotFound))
			})
		})

		Describe("when habit is not found", func() {
			It("has status 404", func() {
				res := api.MustGet(app.Token, "this will not be found")
				Expect(res.StatusCode).To(Equal(http.StatusNotFound))
			})
		})

		It("cannot read habits from other users", func() {
			otherUser := app.MustCreateToken()
			defaultUserHabit := app.MustCreate("read")

			res := api.MustGet(otherUser, defaultUserHabit.ID)

			Expect(res.StatusCode).To(Equal(http.StatusNotFound))
		})

		Describe("activities", func() {
			It("requires token", func() {
				h := app.MustCreate("read")
				res := api.MustAddActivity("token", h.ID, "desc", time.Now())
				Expect(res.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			Describe("create", func() {
				It("returns created activities on get", func() {
					h := app.MustCreate("read")
					app.MustAddActivity(h.ID, time.Now())
					app.MustAddActivity(h.ID, time.Now())
					app.MustAddActivity(h.ID, time.Now())

					h = app.MustGet(h.ID)
					Expect(h.Activities).To(HaveLen(3))
				})

				It("cannot create activities for other user's habits", func() {
					otherUser := app.MustCreateToken()
					defaultUserHabit := app.MustCreate("read")

					res := api.MustAddActivity(otherUser, defaultUserHabit.ID, "desc", time.Now())

					Expect(res.StatusCode).To(Equal(http.StatusNotFound))
				})

				It("contains a description", func() {
					h := app.MustCreate("read")
					app.MustAddActivityWithDesc(h.ID, "my description", time.Now())

					habit := app.MustGet(h.ID)

					Expect(habit.Activities[0].Desc).To(Equal("my description"))
				})
			})

			Describe("update", func() {
				It("updates the description", func() {
					hab := app.MustCreate("read")
					app.MustAddActivityWithDesc(hab.ID, "old", time.Now())

					activity := app.MustGet(hab.ID).Activities[0]
					app.MustUpdateActivityDesc(hab.ID, activity.ID, "new")

					habit := app.MustGet(hab.ID)
					Expect(habit.Activities[0].Desc).To(Equal("new"))
				})

				It("requires token", func() {
					hab := app.MustCreate("read")
					app.MustAddActivityWithDesc(hab.ID, "old", time.Now())
					activity := app.MustGet(hab.ID).Activities[0]

					res := api.MustUpdateActivity("bad token", hab.ID, activity.ID, "new")
					Expect(res.StatusCode).To(Equal(http.StatusUnauthorized))
				})

				It("cannot update activities for other user's activities", func() {
					otherUser := app.MustCreateToken()
					defaultUserHabit := app.MustCreate("read")
					app.MustAddActivityWithDesc(defaultUserHabit.ID, "old", time.Now())
					defaultUserActivity := app.MustGet(defaultUserHabit.ID).Activities[0]

					res := api.MustUpdateActivity(otherUser, defaultUserHabit.ID, defaultUserActivity.ID, "new")

					Expect(res.StatusCode).To(Equal(http.StatusNotFound))
				})

				It("returns 404 when activity not found", func() {
					h := app.MustCreate("read")
					res := api.MustUpdateActivity(app.Token, h.ID, badActivityID, "desc")
					Expect(res.StatusCode).To(Equal(http.StatusNotFound))
				})

				It("returns 404 when given a bad habit id", func() {
					res := api.MustUpdateActivity(app.Token, "not an uuid", "cc4f532a-4076-4dba-ac73-f003ee59ea07", "desc")
					Expect(res.StatusCode).To(Equal(http.StatusNotFound))
				})

				It("returns 404 when given a bad activity id", func() {
					res := api.MustUpdateActivity(app.Token, "cc4f532a-4076-4dba-ac73-f003ee59ea07", "not an uuid", "desc")
					Expect(res.StatusCode).To(Equal(http.StatusNotFound))
				})
			})

			Describe("activity delete", func() {
				It("requires token", func() {
					res := api.MustDeleteActivity("bad token", "habit id", "activity id")
					Expect(res.StatusCode).To(Equal(http.StatusUnauthorized))
				})

				It("is not returned", func() {
					habit := app.MustCreate("read")
					app.MustAddActivity(habit.ID, time.Now())
					habit = app.MustGet(habit.ID)
					Expect(habit.Activities).To(HaveLen(1))

					app.MustDeleteActivity(habit.ID, habit.Activities[0].ID)

					habit = app.MustGet(habit.ID)
					Expect(habit.Activities).To(HaveLen(0))
				})

				It("returns 404 when habit not found", func() {
					habit := app.MustCreate("read")
					app.MustAddActivity(habit.ID, time.Now())
					habit = app.MustGet(habit.ID)
					Expect(habit.Activities).To(HaveLen(1))

					res := api.MustDeleteActivity(app.Token, badHabitID, habit.Activities[0].ID)
					Expect(res.StatusCode).To(Equal(http.StatusNotFound))
				})

				It("returns 404 when activity not found", func() {
					h := app.MustCreate("read")
					res := api.MustDeleteActivity(app.Token, h.ID, badActivityID)
					Expect(res.StatusCode).To(Equal(http.StatusNotFound))
				})

				It("returns 404 when given a bad habit id", func() {
					res := api.MustDeleteActivity(app.Token, "not an uuid", "cc4f532a-4076-4dba-ac73-f003ee59ea07")
					Expect(res.StatusCode).To(Equal(http.StatusNotFound))
				})

				It("returns 404 when given a bad activity id", func() {
					res := api.MustDeleteActivity(app.Token, "cc4f532a-4076-4dba-ac73-f003ee59ea07", "not an uuid")
					Expect(res.StatusCode).To(Equal(http.StatusNotFound))
				})
			})
		})

		Describe("delete", func() {
			It("requires token", func() {
				h := app.MustCreate("read")
				res := api.MustDelete("", h.ID)
				Expect(res.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("return status ok", func() {
				h := app.MustCreate("read")
				res := api.MustDelete(app.Token, h.ID)
				Expect(res.StatusCode).To(Equal(http.StatusOK))
			})

			It("cannot delete other user's habits", func() {
				otherUser := app.MustCreateToken()
				defaultUserHabit := app.MustCreate("read")

				res := api.MustDelete(otherUser, defaultUserHabit.ID)

				Expect(res.StatusCode).To(Equal(http.StatusNotFound))
			})
		})
	})

	Describe("groups", func() {
		Describe("create", func() {
			It("creates a group", func() {
				app.MustCreateGroup("health")
			})

			It("is listed", func() {
				health := app.MustCreateGroup("health")
				groups, _ := app.MustGroupsAndHabits()

				Expect(groups).To(ContainElement(health))
			})
		})

		Describe("habits", func() {
			When("in the group", func() {
				It("is listed in groups", func() {
					health := app.MustCreateGroup("health")
					run := app.MustCreate("run")

					app.MustAddToGroup(run, health)

					groups, _ := app.MustGroupsAndHabits()
					Expect(groups).To(HaveLen(1))
					Expect(groups[0].Habits).To(ContainElement(run))
				})

				It("is not listed in habits", func() {
					health := app.MustCreateGroup("health")
					run := app.MustCreate("run")

					app.MustAddToGroup(run, health)

					_, habits := app.MustGroupsAndHabits()
					Expect(habits).To(BeEmpty())
				})
			})

			When("out of the group", func() {
				It("is not listed in groups", func() {
					app.MustCreateGroup("health")
					app.MustCreate("run")

					groups, _ := app.MustGroupsAndHabits()
					Expect(groups).To(HaveLen(1))
					Expect(groups[0].Habits).To(BeEmpty())
				})

				It("is listed in habits", func() {
					app.MustCreateGroup("health")
					run := app.MustCreate("run")

					_, habits := app.MustGroupsAndHabits()
					Expect(habits).To(ContainElement(run))
				})
			})

			When("removing a habit from the group", func() {
				It("changes it from 'groups' to 'habits'", func() {
					health := app.MustCreateGroup("health")
					run := app.MustCreate("run")

					app.MustAddToGroup(run, health)

					groups, habits := app.MustGroupsAndHabits()
					Expect(groups[0].Habits).To(ContainElement(run))
					Expect(habits).To(BeEmpty())

					app.MustRemoveFromGroup(run, health)

					groups, habits = app.MustGroupsAndHabits()
					Expect(groups[0].Habits).To(BeEmpty())
					Expect(habits).To(ContainElement(run))
				})
			})

			When("removing deleting a group", func() {
				It("changes it from 'groups' to 'habits'", func() {
					health := app.MustCreateGroup("health")
					run := app.MustCreate("run")

					app.MustAddToGroup(run, health)

					groups, habits := app.MustGroupsAndHabits()
					Expect(groups[0].Habits).To(ContainElement(run))
					Expect(habits).To(BeEmpty())

					app.MustDeleteGroup(health)

					groups, habits = app.MustGroupsAndHabits()
					Expect(groups).To(BeEmpty())
					Expect(habits).To(ContainElement(run))
				})
			})
		})
	})
})
