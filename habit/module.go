package habit

import "go.uber.org/fx"

var Module = fx.Module(
	"habit",
	fx.Provide(NewController),

	fx.Provide(NewPromProbe),
	fx.Provide(func(probe *PromProbe) Probe { return probe }),

	fx.Provide(NewHabitService),
	fx.Provide(NewHabitSQLRepository),
	fx.Provide(func(repo *HabitSQLRepository) HabitRepository { return repo }),

	fx.Provide(NewActivityService),
	fx.Provide(NewActivitySQLRepository),
	fx.Provide(func(repo *ActivitySQLRepository) ActivityRepository { return repo }),

	fx.Provide(NewGroupService),
	fx.Provide(NewGroupSQLRepository),
	fx.Provide(func(repo *GroupSQLRepository) GroupRepository { return repo }),
)
