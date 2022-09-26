package token

import (
	"encoding/base64"
)

type Base64Encoder struct{}

func NewBase64Encoder() Base64Encoder {
	return Base64Encoder{}
}

func (e Base64Encoder) Encode(in []byte) ([]byte, error) {
	b := make([]byte, base64.StdEncoding.EncodedLen(len(in)))
	base64.StdEncoding.Encode(b, in)
	return b, nil
}

func (e Base64Encoder) Decode(in []byte) ([]byte, error) {
	b := make([]byte, base64.StdEncoding.DecodedLen(len(in))-1)
	_, err := base64.StdEncoding.Decode(b, in)
	return b, err
}
