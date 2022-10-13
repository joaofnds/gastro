package token

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewService),
	fx.Provide(NewPostgresIDGenerator),
	fx.Provide(NewAceEncrypter),
	fx.Provide(NewBase64Encoder),
	fx.Provide(NewPromInstrumentation),
	fx.Provide(func(idGen *PostgresIDGenerator) IDGenerator { return idGen }),
	fx.Provide(func(encrypter *AceEncrypter) Encrypter { return encrypter }),
	fx.Provide(func(encoder *Base64Encoder) Encoder { return encoder }),
	fx.Provide(func(instr *PromInstrumentation) Instrumentation { return instr }),
)
