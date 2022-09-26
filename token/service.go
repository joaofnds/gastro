package token

type IDGenerator interface {
	NewID() ([]byte, error)
}

type Encrypter interface {
	Encrypt([]byte) ([]byte, error)
	Decrypt([]byte) ([]byte, error)
}

type Encoder interface {
	Encode([]byte) ([]byte, error)
	Decode([]byte) ([]byte, error)
}

type TokenService struct {
	idService IDGenerator
	crypto    Encrypter
	base64    Encoder
}

func NewTokenService(repo *UserIDService, enc *EncryptionService) *TokenService {
	return &TokenService{repo, enc, NewBase64Encoder()}
}

func (t *TokenService) NewToken() ([]byte, error) {
	id, err := t.idService.NewID()
	if err != nil {
		return id, err
	}

	encrypted, err := t.crypto.Encrypt(id)
	if err != nil {
		return id, err
	}

	return t.base64.Encode(encrypted)
}

func (t *TokenService) IdFromToken(token []byte) ([]byte, error) {
	b, err := t.base64.Decode(token)
	if err != nil {
		return b, err
	}

	return t.crypto.Decrypt(b)
}
