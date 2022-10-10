package habit

import "go.uber.org/fx"

var Module = fx.Module(
	"habit",
	fx.Provide(NewHabitService),
	fx.Provide(NewSQLHabitRepository),
	fx.Provide(NewPromHabitInstrumentation),
	fx.Provide(func(repo *SQLHabitRepository) HabitRepository {
		return repo
	}),
	fx.Provide(func(instr *PromHabitInstrumentation) HabitInstrumentation {
		return instr
	}),
)
