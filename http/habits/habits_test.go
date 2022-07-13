package habits_test

import (
	"astro/config"
	"astro/habit"
	"astro/http/fiber"
	"astro/http/habits"
	"astro/postgres"
	"astro/test"
	"net/http"
	"testing"

	"astro/test/driver"
	. "astro/test/matchers"
	"astro/test/transaction"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx/fxtest"
)

func TestHealth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "/habits suite")
}

var _ = Describe("/habits", func() {
	var app *fxtest.App

	BeforeEach(func() {
		app = fxtest.New(
			GinkgoT(),
			test.NopLogger,
			config.Module,
			fiber.Module,
			postgres.Module,
			habit.Module,
			habits.Providers,
			transaction.Module,
		)
		app.RequireStart()
	})

	AfterEach(func() {
		app.RequireStop()
	})

	Describe("GET", func() {
		It("returns a list of habits", func() {
			app := driver.NewDriver()
			Must(app.Create("read"))

			data := Must2(app.List())

			Expect(data).To(HaveLen(1))
			Expect(data[0].Name).To(Equal("read"))
		})
	})

	Describe("POST", func() {
		It("returns status created", func() {
			res, _ := driver.NewAPI().Create("read")
			Expect(res.StatusCode).To(Equal(http.StatusCreated))
		})

		Describe("without name", func() {
			It("return bad request", func() {
				res, _ := driver.NewAPI().Create("")
				Expect(res.StatusCode).To(Equal(http.StatusBadRequest))
			})
		})
	})

	Describe("/:name", func() {
		Describe("when habit is found", func() {
			It("has status 200", func() {
				api := driver.NewAPI()
				Must2(api.Create("read"))

				res := Must2(api.Get("read"))
				Expect(res.StatusCode).To(Equal(http.StatusOK))
			})

			It("returns the habit", func() {
				app := driver.NewDriver()
				Must(app.Create("read"))

				habit := Must2(app.Get("read"))
				Expect(habit.ID > 0).To(BeTrue())
				Expect(habit.Name).To(Equal("read"))
				Expect(habit.Activities).To(HaveLen(0))
			})
		})

		Describe("after deleting the habit", func() {
			It("has status 404", func() {
				api := driver.NewAPI()
				Must2(api.Create("read"))

				res := Must2(api.Get("read"))
				Expect(res.StatusCode).To(Equal(http.StatusOK))

				res = Must2(api.Delete("read"))
				Expect(res.StatusCode).To(Equal(http.StatusOK))

				res = Must2(api.Get("read"))
				Expect(res.StatusCode).To(Equal(http.StatusNotFound))
			})
		})

		Describe("when habit is not found", func() {
			It("has status 404", func() {
				api := driver.NewAPI()
				Must2(api.Create("read"))

				res := Must2(api.Get("not read"))
				Expect(res.StatusCode).To(Equal(http.StatusNotFound))
			})
		})

		Describe("activities", func() {
			It("POST to add acitity", func() {
				app := driver.NewDriver()
				api := driver.NewAPI()
				Must2(api.Create("read"))

				Must2(api.AddActivity("read"))
				Must2(api.AddActivity("read"))
				Must2(api.AddActivity("read"))

				habit := Must2(app.Get("read"))
				Expect(habit.Activities).To(HaveLen(3))
			})
		})
	})
})
