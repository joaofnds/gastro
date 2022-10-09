package metrics_test

import (
	"astro/config"
	. "astro/http/req"
	"astro/metrics"
	"astro/test"
	. "astro/test/matchers"
	"fmt"
	"io"
	"net/http"
	"testing"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHealth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "/health suite")
}

var _ = Describe("/", Ordered, func() {
	var url string

	BeforeAll(func() {
		var metricsConfig config.MetricsConfig
		fxtest.New(
			GinkgoT(),
			test.NopLogger,
			config.Module,
			metrics.Module,
			fx.Populate(&metricsConfig),
		).RequireStart()
		url = fmt.Sprintf("http://%s/metrics", metricsConfig.Address)
	})

	It("returns status OK", func() {
		res := Must2(Get(url, nil))
		Expect(res.StatusCode).To(Equal(http.StatusOK))
	})

	It("returns how many requests were made", func() {
		res := Must2(Get(url, nil))
		b := Must2(io.ReadAll(res.Body))
		Expect(b).To(ContainSubstring(`promhttp_metric_handler_requests_total{code="200"} 1`))
	})
})
