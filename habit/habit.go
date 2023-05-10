package habit

import "go.uber.org/fx"

var Module = fx.Module(
	"habit",
	fx.Provide(NewController),
	fx.Provide(NewService),
	fx.Provide(NewSQLRepository),
	fx.Provide(NewPromProbe),
	fx.Provide(func(repo *SQLRepository) Repository { return repo }),
	fx.Provide(func(probe *PromProbe) Probe { return probe }),
)
