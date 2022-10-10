package health

import (
	. "astro/health"

	"go.uber.org/fx"
)

var UnhealthyHealthService = fx.Decorate(NewUnhealthyHealthService)

func NewUnhealthyHealthService() HealthChecker {
	return &unhealthyHealthService{}
}

type unhealthyHealthService struct{}

func (c *unhealthyHealthService) CheckHealth() HealthCheck {
	return HealthCheck{
		DB: Status{Status: StatusDown},
	}
}
