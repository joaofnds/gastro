package token

import (
	"astro/config"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"filippo.io/age"
	"go.uber.org/fx"
)

var Module = fx.Provide(NewTokenService)

type TokenService struct {
	recipient age.Recipient
	identity  age.Identity
}

func NewTokenService(config config.AppConfig) (*TokenService, error) {
	recipient, err := age.ParseX25519Recipient(config.Token.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse recipient: %w", err)
	}

	identity, err := age.ParseX25519Identity(config.Token.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse identity: %w", err)
	}

	return &TokenService{recipient: recipient, identity: identity}, nil
}

func (t TokenService) Generate(payload string) (string, error) {
	b := new(bytes.Buffer)
	w, err := age.Encrypt(b, t.recipient)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt: %w", err)
	}

	if _, err = io.WriteString(w, payload); err != nil {
		return "", fmt.Errorf("failed to write payload: %w", err)
	}

	if err := w.Close(); err != nil {
		return "", fmt.Errorf("Failed to close encrypted file: %w", err)
	}

	return base64.StdEncoding.EncodeToString(b.Bytes()), nil
}

func (t TokenService) Decode(token string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", fmt.Errorf("failed to base64 decode: %w", err)
	}

	r, err := age.Decrypt(bytes.NewReader(b), t.identity)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	s := new(strings.Builder)
	if _, err := io.Copy(s, r); err != nil {
		return "", fmt.Errorf("failed to read decrypted data: %w", err)
	}

	return s.String(), nil
}
