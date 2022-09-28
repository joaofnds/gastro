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

type TokenInstrumentation interface {
	TokenCreated()
	TokenDecrypted()
	FailedToDecryptToken(err error)
	FailedToCreateToken(err error)
}

type TokenService struct {
	idService IDGenerator
	crypto    Encrypter
	base64    Encoder
	instr     TokenInstrumentation
}

func NewTokenService(repo *UserIDService, enc *EncryptionService, instr TokenInstrumentation) *TokenService {
	return &TokenService{repo, enc, NewBase64Encoder(), instr}
}

func (t *TokenService) NewToken() ([]byte, error) {
	id, err := t.idService.NewID()
	if err != nil {
		t.instr.FailedToCreateToken(err)
		return id, err
	}

	encrypted, err := t.crypto.Encrypt(id)
	if err != nil {
		t.instr.FailedToCreateToken(err)
		return id, err
	}

	tok, err := t.base64.Encode(encrypted)
	if err != nil {
		t.instr.FailedToCreateToken(err)
		return nil, err
	}

	t.instr.TokenCreated()

	return tok, nil
}

func (t *TokenService) IdFromToken(token []byte) ([]byte, error) {
	b, err := t.base64.Decode(token)
	if err != nil {
		t.instr.FailedToDecryptToken(err)
		return b, err
	}

	tok, err := t.crypto.Decrypt(b)
	if err != nil {
		t.instr.FailedToDecryptToken(err)
		return nil, err
	}

	t.instr.TokenDecrypted()
	return tok, nil
}
