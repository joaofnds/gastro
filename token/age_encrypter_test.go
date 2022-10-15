package token_test

import (
	"astro/config"
	"astro/test"
	. "astro/test/matchers"
	"astro/token"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

var _ = Describe("age encrypter", Ordered, func() {
	var service token.Encrypter
	var cfg config.Token

	BeforeAll(func() {
		fxtest.New(
			GinkgoT(),
			test.NopLogger,
			test.NopTokenInstrumentation,
			config.Module,
			token.Module,
			fx.Populate(&cfg, &service),
		)
	})

	It("generates strings that can be parsed back to the same payload", func() {
		payload := []byte("Hello, World!")

		token := Must2(service.Encrypt(payload))
		Expect(token).NotTo(Equal(payload))

		decoded := Must2(service.Decrypt(token))
		Expect(decoded).To(Equal(payload))
	})

	It("does not generate the same string for the same payload", func() {
		payload := []byte("Hello, World!")

		token1 := Must2(service.Encrypt(payload))
		token2 := Must2(service.Encrypt(payload))

		Expect(token1).NotTo(Equal(token2))
	})

	Describe("with invalid public key", func() {
		It("fails to create service", func() {
			_, err := token.NewAgeEncrypter(config.Token{
				PublicKey:  "this is invalid",
				PrivateKey: cfg.PrivateKey,
			})

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to parse recipient"))
		})
	})

	Describe("invalid private key", func() {
		It("fails to create service", func() {
			_, err := token.NewAgeEncrypter(config.Token{
				PublicKey:  cfg.PublicKey,
				PrivateKey: "this is invalid",
			})

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to parse identity"))
		})
	})
})
