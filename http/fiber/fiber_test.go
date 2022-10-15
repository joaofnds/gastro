package fiber_test

import (
	"astro/config"
	"astro/http/fiber"
	"astro/test"
	"fmt"
	"net/http"
	"testing"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	. "astro/test/matchers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

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
		var cfg config.App

		fxApp = fxtest.New(
			GinkgoT(),
			test.NopLogger,
			test.NewPortAppConfig,
			test.NopHTTPInstrumentation,
			test.PanicHandler,
			config.Module,
			fiber.Module,
			fx.Populate(&cfg),
		).RequireStart()

		url = fmt.Sprintf("http://localhost:%d", cfg.Port)
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
