package habit

import (
	"go.uber.org/fx"
)

var NopProbeProvider = fx.Decorate(func() Probe { return NopProbe{} })

type NopProbe struct{}

func (p NopProbe) LogHabitCreated()             {}
func (p NopProbe) LogFailedToCreateHabit(error) {}
