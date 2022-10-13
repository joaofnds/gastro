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

type Instrumentation interface {
	TokenCreated()
	TokenDecrypted()
	FailedToDecryptToken(err error)
	FailedToCreateToken(err error)
}

type Service struct {
	idGen           IDGenerator
	encrypter       Encrypter
	encoder         Encoder
	instrumentation Instrumentation
}

func NewService(id IDGenerator, encrypter Encrypter, encoder Encoder, instrumentation Instrumentation) *Service {
	return &Service{id, encrypter, encoder, instrumentation}
}

func (t *Service) NewToken() ([]byte, error) {
	id, err := t.idGen.NewID()
	if err != nil {
		t.instrumentation.FailedToCreateToken(err)
		return nil, err
	}

	encrypted, err := t.encrypter.Encrypt(id)
	if err != nil {
		t.instrumentation.FailedToCreateToken(err)
		return nil, err
	}

	tok, err := t.encoder.Encode(encrypted)
	if err != nil {
		t.instrumentation.FailedToCreateToken(err)
		return nil, err
	}

	t.instrumentation.TokenCreated()

	return tok, nil
}

func (t *Service) IDFromToken(token []byte) ([]byte, error) {
	b, err := t.encoder.Decode(token)
	if err != nil {
		t.instrumentation.FailedToDecryptToken(err)
		return nil, err
	}

	tok, err := t.encrypter.Decrypt(b)
	if err != nil {
		t.instrumentation.FailedToDecryptToken(err)
		return nil, err
	}

	t.instrumentation.TokenDecrypted()

	return tok, nil
}
