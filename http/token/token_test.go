package token_test

import (
	"astro/config"
	"astro/http/fiber"
	httpToken "astro/http/token"
	"astro/postgres"
	"astro/test"
	"astro/token"
	"io"
	"net/http"
	"testing"

	"astro/test/driver"
	. "astro/test/matchers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx/fxtest"
)

func TestTokenHTTP(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "/token suite")
}

var _ = Describe("/token", Ordered, func() {
	var app *fxtest.App
	api := driver.NewAPI()

	BeforeAll(func() {
		app = fxtest.New(
			GinkgoT(),
			test.NopLogger,
			test.FakeInstrumentation,
			config.Module,
			fiber.Module,
			postgres.Module,
			token.Module,
			httpToken.Providers,
		)
		app.RequireStart()
	})

	AfterAll(func() { app.RequireStop() })

	It("returns status created", func() {
		res := Must2(api.CreateToken())
		Expect(res.StatusCode).To(Equal(http.StatusCreated))
	})

	It("returns the token", func() {
		res := Must2(api.CreateToken())
		body := Must2(io.ReadAll(res.Body))
		defer res.Body.Close()

		Expect(body).To(HaveLen(316))
	})

	Describe("token check", func() {
		It("returns ok for tokens emitted by us", func() {
			res := Must2(api.CreateToken())
			token := Must2(io.ReadAll(res.Body))
			defer res.Body.Close()

			res = Must2(api.TestToken(token))
			Expect(res.StatusCode).To(Equal(http.StatusOK))
		})

		It("returns bad request for invalid tokens", func() {
			res := Must2(api.TestToken([]byte("invalid token")))
			Expect(res.StatusCode).To(Equal(http.StatusBadRequest))
		})
	})
})
