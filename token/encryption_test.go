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

var _ = Describe("encryption service", Ordered, func() {
	var service *token.EncryptionService
	var cfg config.AppConfig

	BeforeAll(func() {
		fxtest.New(
			GinkgoT(),
			test.NopLogger,
			test.NopTokenInstrumentation,
			config.Module,
			fx.Populate(&cfg),
			token.Module,
			fx.Populate(&service),
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
			_, err := token.NewEncryptionService(config.AppConfig{
				Token: config.TokenConfig{
					PublicKey:  "this is invalid",
					PrivateKey: cfg.Token.PrivateKey,
				},
			})

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to parse recipient"))
		})
	})

	Describe("invalid private key", func() {
		It("fails to create service", func() {
			_, err := token.NewEncryptionService(config.AppConfig{
				Token: config.TokenConfig{
					PublicKey:  cfg.Token.PublicKey,
					PrivateKey: "this is invalid",
				},
			})

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to parse identity"))
		})
	})
})
