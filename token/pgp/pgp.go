package pgp

import (
	"bytes"
	"errors"
	"io"

	"golang.org/x/crypto/openpgp"
)

func Encrypt(publickey io.Reader, text []byte) ([]byte, error) {
	if publickey == nil {
		return nil, errors.New("invalid argument")
	}

	key, err := openpgp.ReadArmoredKeyRing(publickey)
	if err != nil {
		return nil, err
	}

	ciphertext := new(bytes.Buffer)
	plaintext, err := openpgp.Encrypt(ciphertext, key, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	if _, err = plaintext.Write(text); err != nil {
		return nil, err
	}

	plaintext.Close()

	return io.ReadAll(ciphertext)
}

func Decrypt(privateKey io.Reader, passphrase string, text []byte) ([]byte, error) {
	if privateKey == nil {
		return nil, errors.New("missing private key")
	}

	keys, err := keys(privateKey, []byte(passphrase))
	if err != nil {
		return nil, err
	}

	messageDetails, err := openpgp.ReadMessage(bytes.NewBuffer(text), keys, nil, nil)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(messageDetails.UnverifiedBody)
}

func keys(privateKey io.Reader, pass []byte) (openpgp.EntityList, error) {
	var keys, err = openpgp.ReadArmoredKeyRing(privateKey)
	if err != nil {
		return nil, err
	}

	key := keys[0]
	key.PrivateKey.Decrypt(pass)

	for _, subkey := range key.Subkeys {
		subkey.PrivateKey.Decrypt(pass)
	}

	return keys, nil
}
