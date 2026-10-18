package http_test

import (
	"astro/adapters/health"
	http2 "astro/adapters/http"
	"astro/adapters/logger"
	"astro/adapters/sqlite"
	"astro/config"
	"astro/habit"
	"astro/test"
	"astro/test/database"
	"astro/token"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
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

var _ = Describe("HTTP middleware and routes", func() {
	var (
		fxApp      *fxtest.App
		httpConfig http2.Config
		module     fx.Option
		url        string
	)

	BeforeEach(func() {
		module = http2.FiberModule
	})

	JustBeforeEach(func() {
		fxApp = fxtest.New(
			GinkgoT(),
			logger.NopLogger,
			test.NewPortAppConfig,
			http2.NopProbeProvider,
			PanicHandler,
			config.Module,
			module,
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

	When("the HTTP module registers application routes", func() {
		BeforeEach(func() {
			module = fx.Options(
				http2.Module,
				health.Module,
				habit.Module,
				token.Module,
				sqlite.Module,
				database.Module,
				habit.NopProbeProvider,
				token.NopProbeProvider,
			)
		})

		It("redirects the root path to the repository", func() {
			client := http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			}}

			res := Must2(client.Get(url + "/"))
			defer func() { _ = res.Body.Close() }()

			Expect(res.StatusCode).To(Equal(http.StatusFound))
			Expect(res.Header.Get("Location")).To(Equal("https://github.com/joaofnds/astro"))
		})
	})
})

var _ = Describe("NewFiber", func() {
	It("leaves the root route unregistered", func() {
		server := http2.NewFiber(http2.Config{Limiter: http2.Limiter{Requests: 120, Expiration: time.Minute}}, http2.NopProbe{})
		DeferCleanup(func() { Expect(server.Shutdown()).To(Succeed()) })

		response := Must2(server.Test(httptest.NewRequest(http.MethodGet, "/", nil)))
		defer func() { _ = response.Body.Close() }()

		Expect(response.StatusCode).To(Equal(http.StatusNotFound))
		Expect(response.Header.Get("Location")).To(BeEmpty())
	})
})

var _ = Describe("HTTP lifecycle", func() {
	It("cancels a queued database operation before waiting for shutdown", func() {
		path := filepath.Join(GinkgoT().TempDir(), "astro.db")
		Expect(sqlite.Initialize(context.Background(), path)).To(Succeed())
		orm, err := sqlite.Open(sqlite.Config{Path: path})
		Expect(err).NotTo(HaveOccurred())
		pool, err := orm.DB()
		Expect(err).NotTo(HaveOccurred())
		held, err := pool.Conn(context.Background())
		Expect(err).NotTo(HaveOccurred())
		server := http2.NewFiber(http2.Config{Limiter: http2.Limiter{Requests: 120, Expiration: time.Minute}}, http2.NopProbe{})
		listening := make(chan fiber.ListenData, 1)
		server.Hooks().OnListen(func(data fiber.ListenData) error { listening <- data; return nil })
		operation := make(chan error, 1)
		deadline := make(chan time.Time, 1)
		server.Get("/queued", func(c fiber.Ctx) error {
			value, _ := c.Context().Deadline()
			deadline <- value
			operation <- pool.PingContext(c.Context())
			return c.SendStatus(fiber.StatusServiceUnavailable)
		})
		application := fx.New(fx.NopLogger, fx.Invoke(func(lifecycle fx.Lifecycle, shutdowner fx.Shutdowner) {
			http2.HookFiber(lifecycle, server, http2.Config{Port: 0}, shutdowner)
		}))
		DeferCleanup(func() {
			_ = held.Close()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			_ = application.Stop(ctx)
			Expect(pool.Close()).To(Succeed())
		})
		Expect(application.Start(context.Background())).To(Succeed())
		var data fiber.ListenData
		Expect(listening).To(Receive(&data))
		clientResult := make(chan error, 1)
		go func() {
			client := http.Client{Timeout: 2 * time.Second}
			response, err := client.Get("http://" + net.JoinHostPort("127.0.0.1", data.Port) + "/queued")
			if err == nil {
				err = response.Body.Close()
			}
			clientResult <- err
		}()
		Eventually(func() int64 { return pool.Stats().WaitCount }).Should(Equal(int64(1)))
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err = application.Stop(ctx)

		Expect(err).NotTo(HaveOccurred())
		Expect(operation).To(Receive(MatchError(context.Canceled)))
		Eventually(clientResult).Should(Receive(BeNil()))
		var bound time.Time
		Expect(deadline).To(Receive(&bound))
		Expect(bound).To(BeTemporally("~", time.Now().Add(30*time.Second), 2*time.Second))
	})

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
