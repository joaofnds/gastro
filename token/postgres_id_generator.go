package token

import (
	"database/sql"
)

type PostgresIDGenerator struct {
	DB Querier
}

type Querier interface {
	QueryRow(string, ...any) *sql.Row
}

func NewPostgresIDGenerator(db *sql.DB) *PostgresIDGenerator {
	return &PostgresIDGenerator{db}
}

func (repo *PostgresIDGenerator) NewID() ([]byte, error) {
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
