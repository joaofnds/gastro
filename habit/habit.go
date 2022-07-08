package habit

import "go.uber.org/fx"

var Module = fx.Options(fx.Provide(NewHabitRepository), fx.Provide(NewHabitService))
