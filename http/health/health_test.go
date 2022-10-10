package health_test

import (
	"astro/config"
	"astro/health"
	"astro/http/fiber"
	httpHealth "astro/http/health"
	"astro/postgres"
	"astro/test"
	testHealth "astro/test/health"
	. "astro/test/matchers"
	"fmt"
	"io"
	"net/http"
	"testing"

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
				fiber.Module,
				httpHealth.Providers,
				fx.Populate(&cfg),
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
				testHealth.UnhealthyHealthService,
				config.Module,
				postgres.Module,
				health.Module,
				fiber.Module,
				httpHealth.Providers,
				fx.Populate(&cfg),
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
