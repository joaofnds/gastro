package health_test

import (
	"astro/config"
	"astro/http/fiber"
	"astro/http/health"
	"astro/test"
	"fmt"
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

var _ = Describe("/", func() {
	var app *fxtest.App
	var cfg config.AppConfig

	BeforeEach(func() {
		app = fxtest.New(
			GinkgoT(),
			test.NopLogger,
			config.Module,
			fiber.Module,
			health.Providers,
			fx.Decorate(test.RandomAppConfigPort),
			fx.Populate(&cfg),
		)
		app.RequireStart()
	})

	AfterEach(func() {
		app.RequireStop()
	})

	It("returns status OK", func() {
		url := fmt.Sprintf("http://localhost:%d/health", cfg.Port)
		res, _ := http.Get(url)
		Expect(res.StatusCode).To(Equal(http.StatusOK))
	})
})
