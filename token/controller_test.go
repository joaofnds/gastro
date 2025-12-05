package token_test

import (
	http2 "astro/adapters/http"
	"astro/adapters/logger"
	"astro/adapters/postgres"
	"astro/config"
	"astro/habit"
	"astro/test"
	"astro/test/driver"
	. "astro/test/matchers"
	"astro/token"
	"fmt"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

var _ = Describe("/token", Ordered, func() {
	var app *fxtest.App
	var api *driver.API

	BeforeAll(func() {
		var httpConfig http2.Config
		app = fxtest.New(
			GinkgoT(),
			logger.NopLogger,
			habit.NopProbeProvider,
			test.NewPortAppConfig,
			http2.NopProbeProvider,
			config.Module,
			http2.FiberModule,
			postgres.Module,
			token.Module,
			fx.Invoke(func(app *fiber.App, controller *token.Controller) {
				controller.Register(app)
			}),
			fx.Populate(&httpConfig),
		)
		app.RequireStart()
		url := fmt.Sprintf("http://localhost:%d", httpConfig.Port)
		api = driver.NewAPI(url)
	})

	AfterAll(func() { app.RequireStop() })

	It("returns status created", func() {
		res := api.MustCreateToken()
		Expect(res.StatusCode).To(Equal(http.StatusCreated))
	})

	It("returns the token", func() {
		res := api.MustCreateToken()
		body := Must2(io.ReadAll(res.Body))
		defer func() { _ = res.Body.Close() }()

		Expect(body).To(HaveLen(316))
	})

	Describe("token check", func() {
		It("returns ok for tokens emitted by us", func() {
			res := api.MustCreateToken()
			token := Must2(io.ReadAll(res.Body))
			defer func() { _ = res.Body.Close() }()

			res = api.MustTestToken(string(token))
			Expect(res.StatusCode).To(Equal(http.StatusOK))
		})

		It("returns bad request for invalid tokens", func() {
			res := api.MustTestToken("invalid token")
			Expect(res.StatusCode).To(Equal(http.StatusBadRequest))
		})
	})
})
