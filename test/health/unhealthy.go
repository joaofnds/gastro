package health

import (
	. "astro/health"

	"go.uber.org/fx"
)

var UnhealthyHealthService = fx.Decorate(NewUnhealthyHealthService)

func NewUnhealthyHealthService() Checker {
	return &unhealthyHealthService{}
}

type unhealthyHealthService struct{}

func (c *unhealthyHealthService) CheckHealth() Check {
	return Check{
		DB: Status{Status: StatusDown},
	}
}
