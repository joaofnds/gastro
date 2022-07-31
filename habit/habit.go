package habit

import "go.uber.org/fx"

var Module = fx.Module(
	"habit",
	fx.Provide(NewHabitRepository),
	fx.Provide(NewHabitService),
	fx.Provide(NewPromHabitInstrumentation),
)
