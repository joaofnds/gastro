package habits_test

import (
	"astro/config"
	"astro/habit"
	"astro/http/fiber"
	"astro/http/habits"
	httpToken "astro/http/token"
	"astro/postgres"
	"astro/test"
	"astro/token"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"astro/test/driver"
	. "astro/test/matchers"
	"astro/test/transaction"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestHabits(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "/habits suite")
}

var _ = Describe("/habits", func() {
	var (
		fxApp *fxtest.App
		app   *driver.Driver
		api   *driver.API
	)

	BeforeEach(func() {
		var cfg config.AppConfig
		fxApp = fxtest.New(
			GinkgoT(),
			test.NopLogger,
			test.RandomAppConfigPort,
			test.NopHabitInstrumentation,
			test.NopTokenInstrumentation,
			config.Module,
			fiber.Module,
			postgres.Module,
			habit.Module,
			habits.Providers,
			transaction.Module,
			token.Module,
			httpToken.Providers,
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

	Describe("/:name", func() {
		It("requires token", func() {
			Must2(app.Create("read"))

			res := Must2(api.Get("", "read"))
			Expect(res.StatusCode).To(Equal(http.StatusUnauthorized))
		})

		Describe("when habit is found", func() {
			It("has status 200", func() {
				Must2(app.Create("read"))

				res := Must2(api.Get(app.Token, "read"))
				Expect(res.StatusCode).To(Equal(http.StatusOK))
			})

			It("returns the habit", func() {
				Must2(app.Create("read"))

				habit := Must2(app.Get("read"))
				Expect(habit.ID > 0).To(BeTrue())
				Expect(habit.Name).To(Equal("read"))
				Expect(habit.Activities).To(HaveLen(0))
			})
		})

		Describe("after deleting the habit", func() {
			It("has status 404", func() {
				api := api
				Must2(api.Create(app.Token, "read"))

				res := Must2(api.Get(app.Token, "read"))
				Expect(res.StatusCode).To(Equal(http.StatusOK))

				res = Must2(api.Delete(app.Token, "read"))
				Expect(res.StatusCode).To(Equal(http.StatusOK))

				res = Must2(api.Get(app.Token, "read"))
				Expect(res.StatusCode).To(Equal(http.StatusNotFound))
			})
		})

		Describe("when habit is not found", func() {
			It("has status 404", func() {
				Must2(api.Create(app.Token, "read"))

				res := Must2(api.Get(app.Token, "not read"))
				Expect(res.StatusCode).To(Equal(http.StatusNotFound))
			})
		})

		It("cannot read habits from other users", func() {
			res := Must2(api.CreateToken())
			defer res.Body.Close()
			otherToken := Must2(io.ReadAll(res.Body))

			Must2(app.Create("read"))

			res = Must2(api.Get(string(otherToken), "read"))

			Expect(res.StatusCode).To(Equal(http.StatusNotFound))
		})

		Describe("activities", func() {
			It("requires token", func() {
				Must2(app.Create("read"))
				res := Must2(api.AddActivity("", "read"))
				Expect(res.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("POST to add activity", func() {
				Must2(app.Create("read"))
				Must(app.AddActivity("read"))
				Must(app.AddActivity("read"))
				Must(app.AddActivity("read"))

				habit := Must2(app.Get("read"))
				Expect(habit.Activities).To(HaveLen(3))
			})
		})

		Describe("delete", func() {
			It("requires token", func() {
				Must2(app.Create("read"))
				res := Must2(api.Delete("", "read"))
				Expect(res.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("return status ok", func() {
				Must2(app.Create("read"))
				res := Must2(api.Delete(app.Token, "read"))
				Expect(res.StatusCode).To(Equal(http.StatusOK))
			})
		})
	})
})
