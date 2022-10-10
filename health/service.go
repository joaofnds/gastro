package health

import (
	"database/sql"
)

type HealthChecker interface {
	CheckHealth() HealthCheck
}

type HealthService struct {
	db *sql.DB
}

func NewHealthService(db *sql.DB) HealthChecker {
	return &HealthService{db}
}

func (c *HealthService) CheckHealth() HealthCheck {
	return HealthCheck{DB: c.DBHealth()}
}

func (c *HealthService) DBHealth() Status {
	if err := c.db.Ping(); err != nil {
		return Status{Status: StatusDown}
	}
	return Status{Status: StatusUp}
}
