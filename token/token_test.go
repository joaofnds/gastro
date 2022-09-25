package token_test

import (
	"encoding/base64"
	"testing"

	"astro/config"
	"astro/test"
	. "astro/test/matchers"
	"astro/token"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestTokenService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Token Service Test")
}

var _ = Describe("token service", Ordered, func() {
	var service *token.TokenService
	var cfg config.AppConfig

	BeforeAll(func() {
		fxtest.New(
			GinkgoT(),
			test.NopLogger,
			config.Module,
			token.Module,
			fx.Populate(&service),
			fx.Populate(&cfg),
		)
	})

	It("generates tokens that can be parsed back to the same payload", func() {
		payload := "Hello, World!"

		token := Must2(service.Generate(payload))
		Expect(token).NotTo(Equal(payload))

		decoded := Must2(service.Decode(token))
		Expect(decoded).To(Equal(payload))
	})

	It("does not generate the same token for the same payload", func() {
		payload := "Hello, World!"

		token1 := Must2(service.Generate(payload))
		token2 := Must2(service.Generate(payload))

		Expect(token1).NotTo(Equal(token2))
	})

	Describe("decode", func() {
		It("fails for non base64 strings", func() {
			_, err := service.Decode("🔑")
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to base64 decode"))
		})

		It("fails to decrypt for plain base64 strings", func() {
			str := base64.StdEncoding.EncodeToString([]byte("✌️"))
			_, err := service.Decode(str)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to decrypt"))
		})
	})

	Describe("with incorrect config", func() {
		Describe("invalid public key", func() {
			It("cannot generate token", func() {
				_, err := token.NewTokenService(config.AppConfig{
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
			It("cannot decode token", func() {
				_, err := token.NewTokenService(config.AppConfig{
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
})
