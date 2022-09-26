package token

import (
	"database/sql"
)

type UserIDService struct {
	DB Querier
}

type Querier interface {
	QueryRow(string, ...any) *sql.Row
}

func NewUserIDService(db *sql.DB) *UserIDService {
	return &UserIDService{db}
}

func (repo *UserIDService) NewID() ([]byte, error) {
	uuid := []byte{}

	row := repo.DB.QueryRow("select uuid_generate_v4()")
	if row.Err() != nil {
		return uuid, row.Err()
	}

	err := row.Scan(&uuid)
	if err != nil {
		return uuid, err
	}

	return uuid, row.Err()
}
