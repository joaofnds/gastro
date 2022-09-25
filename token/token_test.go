package token_test

import (
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

		token := Must2(service.GenToken(payload))
		Expect(token).NotTo(Equal(payload))

		decoded := Must2(service.DecodeToken(token))
		Expect(decoded).To(Equal(payload))
	})

	It("does not generate the same token for the same payload", func() {
		payload := "Hello, World!"

		token1 := Must2(service.GenToken(payload))
		token2 := Must2(service.GenToken(payload))

		Expect(token1).NotTo(Equal(token2))
	})

	Describe("with incorrect config", func() {
		Describe("invalid public key", func() {
			It("cannot generate token", func() {
				badService := token.NewTokenService(config.AppConfig{
					Token: config.TokenConfig{
						PublicKey:  "fuck",
						PrivateKey: cfg.Token.PrivateKey,
						Passphrase: cfg.Token.Passphrase,
					},
				})

				token, err := badService.GenToken("Hello!")
				Expect(token).To(BeEmpty())
				Expect(err).To(MatchError("failed to encrypt: openpgp: invalid argument: no armored data found"))
			})
		})

		Describe("invalid private key", func() {
			It("cannot decode token", func() {
				badService := token.NewTokenService(config.AppConfig{
					Token: config.TokenConfig{
						PublicKey:  cfg.Token.PublicKey,
						PrivateKey: "this is invalid",
						Passphrase: cfg.Token.Passphrase,
					},
				})

				token := Must2(badService.GenToken("Hello!"))

				payload, err := badService.DecodeToken(token)
				Expect(payload).To(BeEmpty())
				Expect(err).To(MatchError("failed to decrypt: openpgp: invalid argument: no armored data found"))
			})
		})

		Describe("invalid passphrase", func() {
			It("cannot decode token", func() {
				badService := token.NewTokenService(config.AppConfig{
					Token: config.TokenConfig{
						PublicKey:  cfg.Token.PublicKey,
						PrivateKey: cfg.Token.PrivateKey,
						Passphrase: "wrong pass",
					},
				})

				token := Must2(badService.GenToken("Hello!"))

				payload, err := badService.DecodeToken(token)
				Expect(payload).To(BeEmpty())
				Expect(err).To(MatchError("failed to decrypt: openpgp: incorrect key"))
			})
		})
	})
})
