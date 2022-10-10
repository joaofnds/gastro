package token

import (
	. "astro/fxutil"

	"go.uber.org/fx"
)

var Module = fx.Options(
	As[IDGenerator](NewPostgresIDGenerator),
	As[Encrypter](NewAceEncrypter),
	As[Encoder](NewBase64Encoder),
	As[TokenInstrumentation](NewPromTokenInstrumentation),
	fx.Provide(NewTokenService),
)
