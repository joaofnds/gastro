package token

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewEncryptionService),
	fx.Provide(NewUserIDService),
	fx.Provide(NewPromTokenInstrumentation),
	fx.Provide(NewTokenService),
)
