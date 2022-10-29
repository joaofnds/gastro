package habit_test

import (
	"astro/config"
	"astro/habit"
	astrohttp "astro/http"
	"astro/postgres"
	"astro/test"
	"astro/test/driver"
	. "astro/test/matchers"
	"astro/test/transaction"
	"astro/token"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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
		var httpConfig astrohttp.Config
		fxApp = fxtest.New(
			GinkgoT(),
			test.NopLogger,
			test.NewPortAppConfig,
			test.NopHabitInstrumentation,
			test.NopTokenInstrumentation,
			test.NopHTTPInstrumentation,
			config.Module,
			astrohttp.FiberModule,
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
			Must2(app.Create("read"))

			data := Must2(app.List())

			Expect(data).To(HaveLen(1))
			Expect(data[0].Name).To(Equal("read"))
		})

		It("requires token", func() {
			res := Must2(api.List(""))
			Expect(res.StatusCode).To(Equal(http.StatusUnauthorized))
		})
	})

	Describe("POST", func() {
		It("returns status created", func() {
			res, _ := api.Create(app.Token, "read")
			Expect(res.StatusCode).To(Equal(http.StatusCreated))
		})

		It("returns the created habit", func() {
			res, _ := api.Create(app.Token, "read")
			body := Must2(io.ReadAll(res.Body))
			defer res.Body.Close()

			var habit habit.Habit
			Must(json.Unmarshal(body, &habit))
			Expect(habit.Name).To(Equal("read"))
		})

		It("requires token", func() {
			res := Must2(api.Create("", "read"))
			Expect(res.StatusCode).To(Equal(http.StatusUnauthorized))
		})

		Describe("without name", func() {
			It("return bad request", func() {
				res, _ := api.Create(app.Token, "")
				Expect(res.StatusCode).To(Equal(http.StatusBadRequest))
			})
		})
	})

	Describe("/:id", func() {
		It("requires token", func() {
			habit := Must2(app.Create("read"))

			res := Must2(api.Get("", habit.ID))
			Expect(res.StatusCode).To(Equal(http.StatusUnauthorized))
		})

		Describe("when habit is found", func() {
			It("has status 200", func() {
				habit := Must2(app.Create("read"))

				res := Must2(api.Get(app.Token, habit.ID))
				Expect(res.StatusCode).To(Equal(http.StatusOK))
			})

			It("returns the habit", func() {
				h := Must2(app.Create("read"))

				habit := Must2(app.Get(h.ID))
				Expect(habit.ID).To(HaveLen(uuidLen))
				Expect(habit.Name).To(Equal("read"))
				Expect(habit.Activities).To(HaveLen(0))
			})
		})

		Describe("after deleting the habit", func() {
			It("has status 404", func() {
				h := Must2(app.Create("read"))

				res := Must2(api.Get(app.Token, h.ID))
				Expect(res.StatusCode).To(Equal(http.StatusOK))

				res = Must2(api.Delete(app.Token, h.ID))
				Expect(res.StatusCode).To(Equal(http.StatusOK))

				res = Must2(api.Get(app.Token, h.ID))
				Expect(res.StatusCode).To(Equal(http.StatusNotFound))
			})
		})

		Describe("when habit is not found", func() {
			It("has status 404", func() {
				res := Must2(api.Get(app.Token, "this will not be found"))
				Expect(res.StatusCode).To(Equal(http.StatusNotFound))
			})
		})

		It("cannot read habits from other users", func() {
			otherUser := Must2(app.CreateToken())
			defaultUserHabit := Must2(app.Create("read"))

			res := Must2(api.Get(otherUser, defaultUserHabit.ID))

			Expect(res.StatusCode).To(Equal(http.StatusNotFound))
		})

		Describe("activities", func() {
			It("requires token", func() {
				h := Must2(app.Create("read"))
				res := Must2(api.AddActivity("token", h.ID, "desc"))
				Expect(res.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			Describe("create", func() {
				It("returns created activites on get", func() {
					h := Must2(app.Create("read"))
					Must(app.AddActivity(h.ID))
					Must(app.AddActivity(h.ID))
					Must(app.AddActivity(h.ID))

					habit := Must2(app.Get(h.ID))
					Expect(habit.Activities).To(HaveLen(3))
				})

				It("cannot create activities for other user's habits", func() {
					otherUser := Must2(app.CreateToken())
					defaultUserHabit := Must2(app.Create("read"))

					res := Must2(api.AddActivity(otherUser, defaultUserHabit.ID, "desc"))

					Expect(res.StatusCode).To(Equal(http.StatusNotFound))
				})

				It("contains a description", func() {
					h := Must2(app.Create("read"))
					Must(app.AddActivityWithDesc(h.ID, "my description"))

					habit := Must2(app.Get(h.ID))

					Expect(habit.Activities[0].Desc).To(Equal("my description"))
				})
			})

			Describe("update", func() {
				It("updates the description", func() {
					hab := Must2(app.Create("read"))
					Must(app.AddActivityWithDesc(hab.ID, "old"))

					activity := Must2(app.Get(hab.ID)).Activities[0]
					Must(app.UpdateActivityDesc(hab.ID, activity.ID, "new"))

					habit := Must2(app.Get(hab.ID))
					Expect(habit.Activities[0].Desc).To(Equal("new"))
				})

				It("requires token", func() {
					hab := Must2(app.Create("read"))
					Must(app.AddActivityWithDesc(hab.ID, "old"))
					activity := Must2(app.Get(hab.ID)).Activities[0]

					res := Must2(api.UpdateActivity("bad token", hab.ID, activity.ID, "new"))
					Expect(res.StatusCode).To(Equal(http.StatusUnauthorized))
				})

				It("cannot update activities for other user's activities", func() {
					otherUser := Must2(app.CreateToken())
					defaultUserHabit := Must2(app.Create("read"))
					Must(app.AddActivityWithDesc(defaultUserHabit.ID, "old"))
					defaultUserActivity := Must2(app.Get(defaultUserHabit.ID)).Activities[0]

					res := Must2(api.UpdateActivity(otherUser, defaultUserHabit.ID, defaultUserActivity.ID, "new"))

					Expect(res.StatusCode).To(Equal(http.StatusNotFound))
				})
			})

			Describe("activity delete", func() {
				It("requires token", func() {
					res := Must2(api.DeleteActivity("bad token", "habit id", "activity id"))
					Expect(res.StatusCode).To(Equal(http.StatusUnauthorized))
				})

				It("is not returned", func() {
					habit := Must2(app.Create("read"))
					Must(app.AddActivity(habit.ID))
					habit = Must2(app.Get(habit.ID))
					Expect(habit.Activities).To(HaveLen(1))

					Must(app.DeleteActivity(habit.ID, habit.Activities[0].ID))

					habit = Must2(app.Get(habit.ID))
					Expect(habit.Activities).To(HaveLen(0))
				})

				It("returns 404 when habit not found", func() {
					habit := Must2(app.Create("read"))
					Must(app.AddActivity(habit.ID))
					habit = Must2(app.Get(habit.ID))
					Expect(habit.Activities).To(HaveLen(1))

					res := Must2(api.DeleteActivity(app.Token, badHabitID, habit.Activities[0].ID))
					Expect(res.StatusCode).To(Equal(http.StatusNotFound))
				})

				It("returns 404 when activity not found", func() {
					h := Must2(app.Create("read"))
					res := Must2(api.DeleteActivity(app.Token, h.ID, badActivityID))
					Expect(res.StatusCode).To(Equal(http.StatusNotFound))
				})

				It("returns 404 when given a bad habit id", func() {
					res := Must2(api.DeleteActivity(app.Token, "not an uuid", "cc4f532a-4076-4dba-ac73-f003ee59ea07"))
					Expect(res.StatusCode).To(Equal(http.StatusNotFound))
				})

				It("returns 404 when given a bad activity id", func() {
					res := Must2(api.DeleteActivity(app.Token, "cc4f532a-4076-4dba-ac73-f003ee59ea07", "not an uuid"))
					Expect(res.StatusCode).To(Equal(http.StatusNotFound))
				})
			})
		})

		Describe("delete", func() {
			It("requires token", func() {
				h := Must2(app.Create("read"))
				res := Must2(api.Delete("", h.ID))
				Expect(res.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("return status ok", func() {
				h := Must2(app.Create("read"))
				res := Must2(api.Delete(app.Token, h.ID))
				Expect(res.StatusCode).To(Equal(http.StatusOK))
			})

			It("cannot delete other user's habits", func() {
				otherUser := Must2(app.CreateToken())
				defaultUserHabit := Must2(app.Create("read"))

				res := Must2(api.Delete(otherUser, defaultUserHabit.ID))

				Expect(res.StatusCode).To(Equal(http.StatusNotFound))
			})
		})
	})
})
