package token

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewTokenService),
	fx.Provide(NewPostgresIDGenerator),
	fx.Provide(NewAceEncrypter),
	fx.Provide(NewBase64Encoder),
	fx.Provide(NewPromTokenInstrumentation),
)
