package health

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewService),
	fx.Provide(func(service *Service) Checker { return service }),
)

const (
	StatusUp   = "up"
	StatusDown = "down"
)

type Checker interface {
	CheckHealth() Check
}
type Check struct {
	DB Status `json:"db"`
}

func (c Check) AllUp() bool {
	return c.DB.IsUp()
}

type Status struct {
	Status string `json:"status"`
}

func (s Status) IsUp() bool {
	return s.Status == StatusUp
}
