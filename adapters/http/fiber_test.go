package http_test

import (
	http2 "astro/adapters/http"
	"astro/adapters/logger"
	"astro/config"
	"astro/test"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v3"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	. "astro/test/matchers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var PanicHandler = fx.Invoke(func(app *fiber.App) {
	app.All("panic", func(c fiber.Ctx) error {
		panic("panic handler")
	})
})

func TestFiber(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "fiber suite")
}

var _ = Describe("fiber middlewares", func() {
	var (
		fxApp      *fxtest.App
		httpConfig http2.Config
		url        string
	)

	BeforeEach(func() {
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

	It("redirects the root path to the repository", func() {
		client := http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}}

		res := Must2(client.Get(url + "/"))

		Expect(res.StatusCode).To(Equal(http.StatusFound))
		Expect(res.Header.Get("Location")).To(Equal("https://github.com/joaofnds/astro"))
	})

	It("reports the requests left in the window", func() {
		res := Must2(http.Get(url + "/somethingelse"))

		Expect(res.Header.Get("X-RateLimit-Limit")).To(Equal(strconv.Itoa(httpConfig.Limiter.Requests)))
		Expect(res.Header.Get("X-RateLimit-Remaining")).To(Equal(strconv.Itoa(httpConfig.Limiter.Requests - 1)))
		Expect(res.Header.Get("X-RateLimit-Reset")).ToNot(BeEmpty())
	})

	It("allows any origin", func() {
		req := Must2(http.NewRequest(http.MethodGet, url+"/somethingelse", nil))
		req.Header.Set("Origin", "https://astro.test")

		res := Must2(http.DefaultClient.Do(req))

		Expect(res.Header.Get("Access-Control-Allow-Origin")).To(Equal("*"))
	})

	It("advertises the methods the service handles", func() {
		req := Must2(http.NewRequest(http.MethodOptions, url+"/somethingelse", nil))
		req.Header.Set("Origin", "https://astro.test")
		req.Header.Set("Access-Control-Request-Method", http.MethodGet)

		res := Must2(http.DefaultClient.Do(req))

		Expect(res.Header.Get("Access-Control-Allow-Methods")).
			To(Equal("GET, POST, HEAD, PUT, DELETE, PATCH"))
	})
})
