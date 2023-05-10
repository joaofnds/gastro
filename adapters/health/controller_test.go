package health_test

import (
	health2 "astro/adapters/health"
	http2 "astro/adapters/http"
	"astro/adapters/logger"
	"astro/adapters/postgres"
	"astro/config"
	"astro/test"
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

var UnhealthyHealthService = fx.Decorate(func() health2.Checker {
	return &unhealthyHealthService{}
})

type unhealthyHealthService struct{}

func (c *unhealthyHealthService) CheckHealth() health2.Check {
	return health2.Check{
		DB: health2.Status{Status: health2.StatusDown},
	}
}

func TestHealth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "/health suite")
}

var _ = Describe("/health", func() {
	var app *fxtest.App
	var url string

	Context("healty", func() {
		BeforeEach(func() {
			var httpConfig http2.Config
			app = fxtest.New(
				GinkgoT(),
				logger.NopLogger,
				test.NewPortAppConfig,
				http2.NopProbeProvider,
				config.Module,
				postgres.Module,
				health2.Module,
				http2.FiberModule,
				fx.Populate(&httpConfig),
				fx.Invoke(func(app *fiber.App, healthController *health2.Controller) {
					healthController.Register(app)
				}),
			)
			url = fmt.Sprintf("http://localhost:%d/health", httpConfig.Port)
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
			var httpConfig http2.Config
			app = fxtest.New(
				GinkgoT(),
				logger.NopLogger,
				test.NewPortAppConfig,
				http2.NopProbeProvider,
				UnhealthyHealthService,
				config.Module,
				postgres.Module,
				health2.Module,
				http2.FiberModule,
				fx.Populate(&httpConfig),
				fx.Invoke(func(app *fiber.App, controller *health2.Controller) {
					controller.Register(app)
				}),
			)
			url = fmt.Sprintf("http://localhost:%d/health", httpConfig.Port)
			app.RequireStart()
		})

		AfterEach(func() { app.RequireStop() })

		It("returns status service unavailable", func() {
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
