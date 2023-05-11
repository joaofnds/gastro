package health

import (
	"astro/adapters/postgres"
	"context"
)

type Service struct {
	postgres *postgres.HealthChecker
}

func NewService(postgres *postgres.HealthChecker) *Service {
	return &Service{postgres}
}

func (s *Service) CheckHealth(ctx context.Context) Check {
	return Check{
		"db": NewStatus(s.postgres.CheckHealth(ctx)),
	}
}
