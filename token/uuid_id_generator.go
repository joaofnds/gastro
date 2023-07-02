package token

import "github.com/google/uuid"

type UUIDGenerator struct{}

func NewUUIDGenerator() *UUIDGenerator {
	return &UUIDGenerator{}
}

func (repo *UUIDGenerator) NewID() ([]byte, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	return []byte(id.String()), nil
}
