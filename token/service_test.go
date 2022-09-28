package token_test

import (
	"astro/config"
	"astro/postgres"
	"astro/test"
	. "astro/test/matchers"
	"astro/token"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

var _ = Describe("token service", Ordered, func() {
	var service *token.TokenService

	BeforeAll(func() {
		fxtest.New(
			GinkgoT(),
			test.NopLogger,
			test.NopTokenInstrumentation,
			config.Module,
			postgres.Module,
			token.Module,
			fx.Populate(&service),
		)
	})

	const (
		tokenLen = 316
		uuidLen  = 36
	)

	It("generates strings that can be parsed back", func() {
		token := Must2(service.NewToken())
		Expect(token).To(HaveLen(tokenLen))

		decoded := Must2(service.IdFromToken(token))
		Expect(decoded).To(HaveLen(uuidLen))
	})
})
