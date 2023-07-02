package token

import "go.uber.org/fx"

var Module = fx.Module(
	"token",
	fx.Provide(NewController),
	fx.Provide(NewService),

	fx.Provide(NewUUIDGenerator),
	fx.Provide(func(gen *UUIDGenerator) IDGenerator { return gen }),

	fx.Provide(NewAgeEncrypter),
	fx.Provide(func(encrypter *AgeEncrypter) Encrypter { return encrypter }),

	fx.Provide(NewBase64Encoder),
	fx.Provide(func(encoder *Base64Encoder) Encoder { return encoder }),

	fx.Provide(NewPromProbe),
	fx.Provide(func(probe *PromProbe) Probe { return probe }),
)
