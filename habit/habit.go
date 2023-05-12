package habit

import "go.uber.org/fx"

var Module = fx.Module(
	"habit",
	fx.Provide(NewController),
	fx.Provide(NewService),

	fx.Provide(NewPromProbe),
	fx.Provide(func(probe *PromProbe) Probe { return probe }),

	fx.Provide(NewSQLHabitRepository),
	fx.Provide(func(repo *SQLHabitRepository) HabitRepository { return repo }),

	fx.Provide(NewSQLActivityRepository),
	fx.Provide(func(repo *SQLActivityRepository) ActivityRepository { return repo }),

	fx.Provide(NewSQLGroupRepository),
	fx.Provide(func(repo *SQLGroupRepository) GroupRepository { return repo }),
)
