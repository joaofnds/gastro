package health

import (
	"database/sql"
)

type Checker interface {
	CheckHealth() Check
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) Checker {
	return &Service{db}
}

func (c *Service) CheckHealth() Check {
	return Check{DB: c.DBHealth()}
}

func (c *Service) DBHealth() Status {
	if err := c.db.Ping(); err != nil {
		return Status{Status: StatusDown}
	}
	return Status{Status: StatusUp}
}
