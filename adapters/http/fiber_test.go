package http_test

import (
	http2 "astro/adapters/http"
	"astro/adapters/logger"
	"astro/config"
	"astro/test"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"testing"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	. "astro/test/matchers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var PanicHandler = fx.Invoke(func(app *fiber.App) {
	app.All("panic", func(c *fiber.Ctx) error {
		panic("panic handler")
	})
})

func TestFiber(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "fiber suite")
}

var _ = Describe("fiber middlewares", func() {
	var (
		fxApp *fxtest.App
		url   string
	)

	BeforeEach(func() {
		var httpConfig http2.Config

		fxApp = fxtest.New(
			GinkgoT(),
			logger.NopLogger,
			test.NewPortAppConfig,
			http2.NopProbeProvider,
			PanicHandler,
			config.Module,
			http2.FiberModule,
			fx.Populate(&httpConfig),
		).RequireStart()

		url = fmt.Sprintf("http://localhost:%d", httpConfig.Port)
	})

	AfterEach(func() {
		fxApp.RequireStop()
	})

	It("recovers from panic", func() {
		req := Must2(http.Get(url + "/panic"))
		Expect(req.StatusCode).To(Equal(http.StatusInternalServerError))

		req = Must2(http.Get(url + "/somethingelse"))
		Expect(req.StatusCode).To(Equal(http.StatusNotFound))
	})

	It("limits requests", func() {
		for i := 0; i < 120; i++ {
			req := Must2(http.Get(url + "/somethingelse"))
			Expect(req.StatusCode).To(Equal(http.StatusNotFound))
		}

		req := Must2(http.Get(url + "/somethingelse"))
		Expect(req.StatusCode).To(Equal(http.StatusTooManyRequests))
	})
})
