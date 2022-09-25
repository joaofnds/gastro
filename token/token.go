package token

import (
	"astro/config"
	"astro/token/pgp"
	"encoding/base64"
	"fmt"
	"strings"

	"go.uber.org/fx"
)

var Module = fx.Provide(NewTokenService)

type TokenService struct {
	pub  string
	priv string
	pass string
}

func NewTokenService(config config.AppConfig) *TokenService {
	return &TokenService{
		pub:  config.Token.PublicKey,
		priv: config.Token.PrivateKey,
		pass: config.Token.Passphrase,
	}
}

func (t TokenService) GenToken(payload string) (string, error) {
	b, err := pgp.Encrypt(strings.NewReader(t.pub), []byte(payload))
	if err != nil {
		return "", fmt.Errorf("failed to encrypt: %w", err)
	}

	token := base64.StdEncoding.EncodeToString(b)
	return token, nil
}

func (t TokenService) DecodeToken(token string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", fmt.Errorf("failed to base64 decode: %w", err)
	}

	b, err = pgp.Decrypt(strings.NewReader(t.priv), t.pass, b)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}
	return string(b), nil
}
