package token_test

import (
	"astro/adapters/logger"
	"astro/config"
	. "astro/test/matchers"
	"astro/token"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

var _ = Describe("age encrypter", Ordered, func() {
	var service token.Encrypter
	var tokenConfig token.Config

	BeforeAll(func() {
		fxtest.New(
			GinkgoT(),
			logger.NopLogger,
			token.NopProbeProvider,
			config.Module,
			token.Module,
			fx.Populate(&tokenConfig, &service),
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
			_, err := token.NewAgeEncrypter(token.Config{
				PublicKey:  "this is invalid",
				PrivateKey: tokenConfig.PrivateKey,
			})

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to parse recipient"))
		})
	})

	Describe("invalid private key", func() {
		It("fails to create service", func() {
			_, err := token.NewAgeEncrypter(token.Config{
				PublicKey:  tokenConfig.PublicKey,
				PrivateKey: "this is invalid",
			})

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to parse identity"))
		})
	})
})
