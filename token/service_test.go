package token_test

import (
	"astro/config"
	"astro/logger"
	"astro/postgres"
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
		service   *token.Service
		idGen     token.IDGenerator
		encrypter token.Encrypter
		encoder   token.Encoder
		probe     token.Probe
	)

	BeforeAll(func() {
		fxtest.New(
			GinkgoT(),
			logger.NopLogger,
			token.NopProbeProvider,
			config.Module,
			postgres.Module,
			token.Module,
			fx.Populate(&service, &idGen, &encrypter, &encoder, &probe),
		)
	})

	const (
		tokenLen = 316
		uuidLen  = 36
	)

	It("generates strings that can be parsed back", func() {
		token := Must2(service.NewToken())
		Expect(token).To(HaveLen(tokenLen))

		decoded := Must2(service.IDFromToken(token))
		Expect(decoded).To(HaveLen(uuidLen))
	})

	Describe("NewToken", func() {

		It("reports that the token was created", func() {
			c := gomock.NewController(GinkgoT())
			defer c.Finish()

			mockProbe := token.NewMockInstrumentation(c)
			mockProbe.EXPECT().TokenCreated()

			service = token.NewService(idGen, encrypter, encoder, mockProbe)
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

				service = token.NewService(mockIdGen, encrypter, encoder, probe)
				_, err := service.NewToken()

				Expect(err).To(Equal(err))
			})

			It("reports the failure", func() {
				c := gomock.NewController(GinkgoT())
				defer c.Finish()

				idGenErr := errors.New("could not generate an id")
				mockIdGen := token.NewMockIDGenerator(c)
				mockIdGen.EXPECT().NewID().Return([]byte{}, idGenErr)

				mockProbe := token.NewMockInstrumentation(c)
				mockProbe.EXPECT().FailedToCreateToken(idGenErr)

				service = token.NewService(mockIdGen, encrypter, encoder, mockProbe)
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

				service = token.NewService(idGen, mockEncrypter, encoder, probe)
				_, err := service.NewToken()

				Expect(err).To(Equal(err))
			})

			It("reports the failure", func() {
				c := gomock.NewController(GinkgoT())
				defer c.Finish()

				encrypterErr := errors.New("could not encrypt the id")
				mockEncrypter := token.NewMockEncrypter(c)
				mockEncrypter.EXPECT().Encrypt(gomock.Any()).Return([]byte{}, encrypterErr)

				mockProbe := token.NewMockInstrumentation(c)
				mockProbe.EXPECT().FailedToCreateToken(encrypterErr)

				service = token.NewService(idGen, mockEncrypter, encoder, mockProbe)
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

				service = token.NewService(idGen, encrypter, mockEncoder, probe)
				_, err := service.NewToken()

				Expect(err).To(Equal(err))
			})

			It("reports the failure", func() {
				c := gomock.NewController(GinkgoT())
				defer c.Finish()

				encoderErr := errors.New("could not encode the id")
				mockEncoder := token.NewMockEncoder(c)
				mockEncoder.EXPECT().Encode(gomock.Any()).Return([]byte{}, encoderErr)

				mockProbe := token.NewMockInstrumentation(c)
				mockProbe.EXPECT().FailedToCreateToken(encoderErr)

				service = token.NewService(idGen, encrypter, mockEncoder, mockProbe)
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

			mockProbe := token.NewMockInstrumentation(c)
			mockProbe.EXPECT().TokenCreated()
			mockProbe.EXPECT().TokenDecrypted()

			service = token.NewService(idGen, encrypter, encoder, mockProbe)
			tok, err := service.NewToken()
			Expect(err).To(BeNil())

			id, err := service.IDFromToken(tok)

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

				service = token.NewService(idGen, encrypter, mockEncoder, probe)
				b, err := service.IDFromToken([]byte{})

				Expect(b).To(BeNil())
				Expect(err).To(Equal(err))
			})

			It("reports the failure", func() {
				c := gomock.NewController(GinkgoT())
				defer c.Finish()

				encoderErr := errors.New("could not decode the id")
				mockEncoder := token.NewMockEncoder(c)
				mockEncoder.EXPECT().Decode(gomock.Any()).Return([]byte{}, encoderErr)

				mockProbe := token.NewMockInstrumentation(c)
				mockProbe.EXPECT().FailedToDecryptToken(encoderErr)

				service = token.NewService(idGen, encrypter, mockEncoder, mockProbe)
				id, err := service.IDFromToken([]byte{})

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

				service = token.NewService(idGen, mockEncrypter, encoder, probe)
				b, err := service.IDFromToken([]byte{})

				Expect(b).To(BeNil())
				Expect(err).To(Equal(err))
			})

			It("reports the failure", func() {
				c := gomock.NewController(GinkgoT())
				defer c.Finish()

				encrypterErr := errors.New("could not decrypt the id")
				mockEncrypter := token.NewMockEncrypter(c)
				mockEncrypter.EXPECT().Decrypt(gomock.Any()).Return([]byte{}, encrypterErr)

				mockProbe := token.NewMockInstrumentation(c)
				mockProbe.EXPECT().FailedToDecryptToken(encrypterErr)

				service = token.NewService(idGen, mockEncrypter, encoder, mockProbe)
				id, err := service.IDFromToken([]byte{})

				Expect(id).To(BeNil())
				Expect(err).To(Equal(encrypterErr))
			})
		})
	})
})
