package token_test

import (
	"astro/config"
	"astro/postgres"
	"astro/test"
	. "astro/test/matchers"
	"astro/token"
	"errors"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

var _ = Describe("token service", Ordered, func() {
	var (
		service         *token.TokenService
		idGen           token.IDGenerator
		encrypter       token.Encrypter
		encoder         token.Encoder
		instrumentation token.TokenInstrumentation
	)

	BeforeAll(func() {
		fxtest.New(
			GinkgoT(),
			test.NopLogger,
			test.NopTokenInstrumentation,
			config.Module,
			postgres.Module,
			token.Module,
			fx.Populate(&service, &idGen, &encrypter, &encoder, &instrumentation),
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

	Describe("NewToken", func() {

		It("reports that the token was created", func() {
			c := gomock.NewController(GinkgoT())
			defer c.Finish()

			mockInstrumentation := token.NewMockTokenInstrumentation(c)
			mockInstrumentation.EXPECT().TokenCreated()

			service = token.NewTokenService(idGen, encrypter, encoder, mockInstrumentation)
			_, err := service.NewToken()
			Expect(err).To(BeNil())
		})

		Describe("when id generator errors", func() {
			It("returns an error", func() {
				c := gomock.NewController(GinkgoT())
				defer c.Finish()

				idGenErr := errors.New("could not generate an id")
				mockIdGen := token.NewMockIDGenerator(c)
				mockIdGen.EXPECT().NewID().Return([]byte{}, idGenErr)

				service = token.NewTokenService(mockIdGen, encrypter, encoder, instrumentation)
				_, err := service.NewToken()

				Expect(err).To(Equal(err))
			})

			It("reports the failure", func() {
				c := gomock.NewController(GinkgoT())
				defer c.Finish()

				idGenErr := errors.New("could not generate an id")
				mockIdGen := token.NewMockIDGenerator(c)
				mockIdGen.EXPECT().NewID().Return([]byte{}, idGenErr)

				mockInstrumentation := token.NewMockTokenInstrumentation(c)
				mockInstrumentation.EXPECT().FailedToCreateToken(idGenErr)

				service = token.NewTokenService(mockIdGen, encrypter, encoder, mockInstrumentation)
				id, err := service.NewToken()
				Expect(id).To(BeNil())
				Expect(err).To(Equal(idGenErr))
			})
		})

		Describe("when encrypter errors", func() {
			It("returns an error", func() {
				c := gomock.NewController(GinkgoT())
				defer c.Finish()

				encrypterErr := errors.New("could not encrypt the id")
				mockEncrypter := token.NewMockEncrypter(c)
				mockEncrypter.EXPECT().Encrypt(gomock.Any()).Return([]byte{}, encrypterErr)

				service = token.NewTokenService(idGen, mockEncrypter, encoder, instrumentation)
				_, err := service.NewToken()

				Expect(err).To(Equal(err))
			})

			It("reports the failure", func() {
				c := gomock.NewController(GinkgoT())
				defer c.Finish()

				encrypterErr := errors.New("could not encrypt the id")
				mockEncrypter := token.NewMockEncrypter(c)
				mockEncrypter.EXPECT().Encrypt(gomock.Any()).Return([]byte{}, encrypterErr)

				mockInstrumentation := token.NewMockTokenInstrumentation(c)
				mockInstrumentation.EXPECT().FailedToCreateToken(encrypterErr)

				service = token.NewTokenService(idGen, mockEncrypter, encoder, mockInstrumentation)
				id, err := service.NewToken()
				Expect(id).To(BeNil())
				Expect(err).To(Equal(encrypterErr))
			})
		})

		Describe("when encoder errors", func() {
			It("returns an error", func() {
				c := gomock.NewController(GinkgoT())
				defer c.Finish()

				encoderErr := errors.New("could not encode the id")
				mockEncoder := token.NewMockEncoder(c)
				mockEncoder.EXPECT().Encode(gomock.Any()).Return([]byte{}, encoderErr)

				service = token.NewTokenService(idGen, encrypter, mockEncoder, instrumentation)
				_, err := service.NewToken()

				Expect(err).To(Equal(err))
			})

			It("reports the failure", func() {
				c := gomock.NewController(GinkgoT())
				defer c.Finish()

				encoderErr := errors.New("could not encode the id")
				mockEncoder := token.NewMockEncoder(c)
				mockEncoder.EXPECT().Encode(gomock.Any()).Return([]byte{}, encoderErr)

				mockInstrumentation := token.NewMockTokenInstrumentation(c)
				mockInstrumentation.EXPECT().FailedToCreateToken(encoderErr)

				service = token.NewTokenService(idGen, encrypter, mockEncoder, mockInstrumentation)
				id, err := service.NewToken()
				Expect(id).To(BeNil())
				Expect(err).To(Equal(encoderErr))
			})
		})
	})

	Describe("IdFromToken", func() {
		It("reports that the token was decrypted", func() {
			c := gomock.NewController(GinkgoT())
			defer c.Finish()

			mockInstrumentation := token.NewMockTokenInstrumentation(c)
			mockInstrumentation.EXPECT().TokenCreated()
			mockInstrumentation.EXPECT().TokenDecrypted()

			service = token.NewTokenService(idGen, encrypter, encoder, mockInstrumentation)
			tok, err := service.NewToken()
			Expect(err).To(BeNil())

			id, err := service.IdFromToken(tok)

			Expect(id).To(HaveLen(uuidLen))
			Expect(err).To(BeNil())
		})

		Describe("when encoder errors", func() {
			It("returns an error", func() {
				c := gomock.NewController(GinkgoT())
				defer c.Finish()

				encoderErr := errors.New("could not decode the id")
				mockEncoder := token.NewMockEncoder(c)
				mockEncoder.EXPECT().Decode(gomock.Any()).Return([]byte{}, encoderErr)

				service = token.NewTokenService(idGen, encrypter, mockEncoder, instrumentation)
				b, err := service.IdFromToken([]byte{})

				Expect(b).To(BeNil())
				Expect(err).To(Equal(err))
			})

			It("reports the failure", func() {
				c := gomock.NewController(GinkgoT())
				defer c.Finish()

				encoderErr := errors.New("could not decode the id")
				mockEncoder := token.NewMockEncoder(c)
				mockEncoder.EXPECT().Decode(gomock.Any()).Return([]byte{}, encoderErr)

				mockInstrumentation := token.NewMockTokenInstrumentation(c)
				mockInstrumentation.EXPECT().FailedToDecryptToken(encoderErr)

				service = token.NewTokenService(idGen, encrypter, mockEncoder, mockInstrumentation)
				id, err := service.IdFromToken([]byte{})

				Expect(id).To(BeNil())
				Expect(err).To(Equal(encoderErr))
			})
		})

		Describe("when encrypter errors", func() {
			It("returns an error", func() {
				c := gomock.NewController(GinkgoT())
				defer c.Finish()

				encrypterErr := errors.New("could not decrypt the id")
				mockEncrypter := token.NewMockEncrypter(c)
				mockEncrypter.EXPECT().Decrypt(gomock.Any()).Return([]byte{}, encrypterErr)

				service = token.NewTokenService(idGen, mockEncrypter, encoder, instrumentation)
				b, err := service.IdFromToken([]byte{})

				Expect(b).To(BeNil())
				Expect(err).To(Equal(err))
			})

			It("reports the failure", func() {
				c := gomock.NewController(GinkgoT())
				defer c.Finish()

				encrypterErr := errors.New("could not decrypt the id")
				mockEncrypter := token.NewMockEncrypter(c)
				mockEncrypter.EXPECT().Decrypt(gomock.Any()).Return([]byte{}, encrypterErr)

				mockInstrumentation := token.NewMockTokenInstrumentation(c)
				mockInstrumentation.EXPECT().FailedToDecryptToken(encrypterErr)

				service = token.NewTokenService(idGen, mockEncrypter, encoder, mockInstrumentation)
				id, err := service.IdFromToken([]byte{})

				Expect(id).To(BeNil())
				Expect(err).To(Equal(encrypterErr))
			})
		})
	})
})
