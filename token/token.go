package token

import "go.uber.org/fx"

var Module = fx.Module(
	"token",
	fx.Provide(NewController),
	fx.Provide(NewService),
	fx.Provide(NewPostgresIDGenerator),
	fx.Provide(NewAgeEncrypter),
	fx.Provide(NewBase64Encoder),
	fx.Provide(NewPromProbe),
	fx.Provide(func(idGen *PostgresIDGenerator) IDGenerator { return idGen }),
	fx.Provide(func(encrypter *AgeEncrypter) Encrypter { return encrypter }),
	fx.Provide(func(encoder *Base64Encoder) Encoder { return encoder }),
	fx.Provide(func(probe *PromProbe) Probe { return probe }),
)
