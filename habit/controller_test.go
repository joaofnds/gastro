package habit_test

import (
	"astro/config"
	"astro/habit"
	astrofiber "astro/http/fiber"
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
		fxApp   *fxtest.App
		app     *driver.Driver
		api     *driver.API
		uuidLen = 36
	)

	BeforeEach(func() {
		var cfg config.App
		fxApp = fxtest.New(
			GinkgoT(),
			test.NopLogger,
			test.NewPortAppConfig,
			test.NopHabitInstrumentation,
			test.NopTokenInstrumentation,
			test.NopHTTPInstrumentation,
			config.Module,
			astrofiber.Module,
			postgres.Module,
			habit.Module,
			transaction.Module,
			token.Module,
			fx.Invoke(func(app *fiber.App, habitController *habit.Controller, tokenController *token.Controller) {
				habitController.Register(app)
				tokenController.Register(app)
			}),
			fx.Populate(&cfg),
		).RequireStart()

		url := fmt.Sprintf("http://localhost:%d", cfg.Port)
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
				res := Must2(api.AddActivity("", h.ID))
				Expect(res.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("POST to add activity", func() {
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

				res := Must2(api.AddActivity(otherUser, defaultUserHabit.ID))

				Expect(res.StatusCode).To(Equal(http.StatusNotFound))
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
