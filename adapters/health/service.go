package health

import (
	"astro/adapters/sqlite"
	"context"
)

type Service struct {
	storage *sqlite.HealthChecker
}

func NewService(storage *sqlite.HealthChecker) *Service {
	return &Service{storage}
}

func (s *Service) CheckHealth(ctx context.Context) Check {
	return Check{
		"db": NewStatus(s.storage.CheckHealth(ctx)),
	}
}
