package token

import (
	"astro/config"
	"bytes"
	"fmt"
	"io"

	"filippo.io/age"
)

type EncryptionService struct {
	recipient age.Recipient
	identity  age.Identity
}

func NewEncryptionService(config config.AppConfig) (*EncryptionService, error) {
	recipient, err := age.ParseX25519Recipient(config.Token.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse recipient: %w", err)
	}

	identity, err := age.ParseX25519Identity(config.Token.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse identity: %w", err)
	}

	return &EncryptionService{recipient: recipient, identity: identity}, nil
}

func (t EncryptionService) Encrypt(data []byte) ([]byte, error) {
	b := new(bytes.Buffer)
	w, err := age.Encrypt(b, t.recipient)
	if err != nil {
		return b.Bytes(), fmt.Errorf("failed to encrypt: %w", err)
	}

	if _, err = w.Write(data); err != nil {
		return b.Bytes(), fmt.Errorf("failed to write payload: %w", err)
	}

	if err := w.Close(); err != nil {
		return b.Bytes(), fmt.Errorf("failed to close encrypted file: %w", err)
	}

	return b.Bytes(), nil
}

func (t EncryptionService) Decrypt(data []byte) ([]byte, error) {
	b := new(bytes.Buffer)

	r, err := age.Decrypt(bytes.NewReader(data), t.identity)
	if err != nil {
		return b.Bytes(), fmt.Errorf("failed to decrypt: %w", err)
	}

	if _, err := io.Copy(b, r); err != nil {
		return b.Bytes(), fmt.Errorf("failed to read decrypted data: %w", err)
	}

	return b.Bytes(), nil
}
