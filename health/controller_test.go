package health_test

import (
	"astro/config"
	"astro/health"
	astrofiber "astro/http/fiber"
	"astro/postgres"
	"astro/test"
	testhealth "astro/test/health"
	. "astro/test/matchers"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestHealth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "/health suite")
}

var _ = Describe("/health", func() {
	var app *fxtest.App
	var url string

	Context("healty", func() {
		BeforeEach(func() {
			var cfg config.AppConfig
			app = fxtest.New(
				GinkgoT(),
				test.NopLogger,
				test.NewPortAppConfig,
				test.NopHTTPInstrumentation,
				config.Module,
				postgres.Module,
				health.Module,
				astrofiber.Module,
				fx.Populate(&cfg),
				fx.Invoke(func(app *fiber.App, healthController *health.Controller) {
					healthController.Register(app)
				}),
			)
			url = fmt.Sprintf("http://localhost:%d/health", cfg.Port)
			app.RequireStart()
		})

		AfterEach(func() { app.RequireStop() })

		It("returns status OK", func() {
			res := Must2(http.Get(url))
			Expect(res.StatusCode).To(Equal(http.StatusOK))
		})

		It("contains info about the db", func() {
			res := Must2(http.Get(url))
			b := Must2(io.ReadAll(res.Body))
			Expect(b).To(ContainSubstring(`"db":{"status":"up"}`))
		})
	})

	Context("unhealty", func() {
		BeforeEach(func() {
			var cfg config.AppConfig
			app = fxtest.New(
				GinkgoT(),
				test.NopLogger,
				test.NewPortAppConfig,
				test.NopHTTPInstrumentation,
				testhealth.UnhealthyHealthService,
				config.Module,
				postgres.Module,
				health.Module,
				astrofiber.Module,
				fx.Populate(&cfg),
				fx.Invoke(func(app *fiber.App, controller *health.Controller) {
					controller.Register(app)
				}),
			)
			url = fmt.Sprintf("http://localhost:%d/health", cfg.Port)
			app.RequireStart()
		})

		AfterEach(func() { app.RequireStop() })

		It("returns status OK", func() {
			res := Must2(http.Get(url))
			Expect(res.StatusCode).To(Equal(http.StatusServiceUnavailable))
		})

		It("contains info about the db", func() {
			res := Must2(http.Get(url))
			b := Must2(io.ReadAll(res.Body))
			Expect(b).To(ContainSubstring(`"db":{"status":"down"}`))
		})
	})
})
