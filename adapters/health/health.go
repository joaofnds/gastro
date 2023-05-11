package health

import "go.uber.org/fx"

var Module = fx.Module(
	"health",
	fx.Provide(NewController),
	fx.Provide(NewService),
	fx.Provide(func(service *Service) Checker { return service }),
)
