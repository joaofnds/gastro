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

type Probe interface {
	TokenCreated()
	TokenDecrypted()
	FailedToDecryptToken(err error)
	FailedToCreateToken(err error)
}

type Service struct {
	idGen     IDGenerator
	encrypter Encrypter
	encoder   Encoder
	probe     Probe
}

func NewService(id IDGenerator, encrypter Encrypter, encoder Encoder, probe Probe) *Service {
	return &Service{id, encrypter, encoder, probe}
}

func (t *Service) NewToken() ([]byte, error) {
	id, err := t.idGen.NewID()
	if err != nil {
		t.probe.FailedToCreateToken(err)
		return nil, err
	}

	encrypted, err := t.encrypter.Encrypt(id)
	if err != nil {
		t.probe.FailedToCreateToken(err)
		return nil, err
	}

	tok, err := t.encoder.Encode(encrypted)
	if err != nil {
		t.probe.FailedToCreateToken(err)
		return nil, err
	}

	t.probe.TokenCreated()

	return tok, nil
}

func (t *Service) IDFromToken(token []byte) ([]byte, error) {
	b, err := t.encoder.Decode(token)
	if err != nil {
		t.probe.FailedToDecryptToken(err)
		return nil, err
	}

	tok, err := t.encrypter.Decrypt(b)
	if err != nil {
		t.probe.FailedToDecryptToken(err)
		return nil, err
	}

	t.probe.TokenDecrypted()

	return tok, nil
}
