package http_test

import (
	http2 "astro/adapters/http"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus"

	. "astro/test/matchers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var promProbe = http2.NewPromProbe()

var _ = Describe("PromProbe", func() {
	var app *fiber.App

	BeforeEach(func() {
		app = fiber.New()
		app.Use(promProbe.Middleware)
		app.Get("/hello", func(c fiber.Ctx) error { return c.SendString("hello") })
	})

	It("forwards the request to the next handler", func() {
		res := Must2(app.Test(httptest.NewRequest(http.MethodGet, "/hello", nil)))

		Expect(res.StatusCode).To(Equal(http.StatusOK))
		Expect(Must2(io.ReadAll(res.Body))).To(Equal([]byte("hello")))
	})

	It("counts the request under the matched route", func() {
		Must2(app.Test(httptest.NewRequest(http.MethodGet, "/hello", nil)))

		Expect(requestSeries()).To(ContainElement(map[string]string{
			"method": http.MethodGet,
			"path":   "/hello",
			"status": "200",
		}))
	})

	When("no route matches", func() {
		It("does not fail the request", func() {
			res := Must2(app.Test(httptest.NewRequest(http.MethodGet, "/nope", nil)))

			Expect(res.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("counts the request as the root route, before the 404 is written", func() {
			Must2(app.Test(httptest.NewRequest(http.MethodGet, "/nope", nil)))

			Expect(requestSeries()).To(ContainElement(map[string]string{
				"method": http.MethodGet,
				"path":   "/",
				"status": "200",
			}))
		})
	})
})

func requestSeries() []map[string]string {
	var series []map[string]string

	for _, family := range Must2(prometheus.DefaultGatherer.Gather()) {
		if family.GetName() != "astro_request" {
			continue
		}

		for _, metric := range family.GetMetric() {
			labels := map[string]string{}
			for _, pair := range metric.GetLabel() {
				labels[pair.GetName()] = pair.GetValue()
			}
			delete(labels, "ip")
			series = append(series, labels)
		}
	}

	return series
}
