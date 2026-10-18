package http_test

import (
	http2 "astro/adapters/http"
	"astro/adapters/logger"
	"astro/config"
	"astro/test"
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

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

var _ = Describe("HTTP lifecycle", func() {
	It("serves requests before startup returns", func() {
		server := fiber.New()
		listening := make(chan fiber.ListenData, 1)
		server.Hooks().OnListen(func(data fiber.ListenData) error { listening <- data; return nil })
		application := fx.New(fx.NopLogger, fx.Invoke(func(lifecycle fx.Lifecycle, shutdowner fx.Shutdowner) {
			http2.HookFiber(lifecycle, server, http2.Config{Port: 0}, shutdowner)
		}))
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		Expect(application.Start(ctx)).To(Succeed())
		DeferCleanup(func() { Expect(application.Stop(context.Background())).To(Succeed()) })

		var data fiber.ListenData
		Expect(listening).To(Receive(&data))
		client := http.Client{Timeout: time.Second}
		response, err := client.Get("http://" + net.JoinHostPort("127.0.0.1", data.Port) + "/missing")

		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = response.Body.Close() }()
		Expect(response.StatusCode).To(Equal(http.StatusNotFound))
	})

	It("returns a startup error when the port is already occupied", func() {
		listener, err := net.Listen("tcp4", ":0")
		Expect(err).NotTo(HaveOccurred())
		defer func() { Expect(listener.Close()).To(Succeed()) }()
		address, ok := listener.Addr().(*net.TCPAddr)
		Expect(ok).To(BeTrue())
		port := address.Port
		server := fiber.New()
		application := fx.New(fx.NopLogger, fx.Invoke(func(lifecycle fx.Lifecycle, shutdowner fx.Shutdowner) {
			http2.HookFiber(lifecycle, server, http2.Config{Port: port}, shutdowner)
		}))
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err = application.Start(ctx)

		Expect(err).To(HaveOccurred())
	})
})
