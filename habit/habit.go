package habit

import "go.uber.org/fx"

var Module = fx.Module(
	"habit",
	fx.Provide(NewService),
	fx.Provide(NewSQLRepository),
	fx.Provide(NewPromInstrumentation),
	fx.Provide(func(repo *SQLRepository) Repository {
		return repo
	}),
	fx.Provide(func(instr *PromInstrumentation) Instrumentation {
		return instr
	}),
)
