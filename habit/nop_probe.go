package habit

import (
	"go.uber.org/fx"
)

var _ Probe = (*NopProbe)(nil)

var NopProbeProvider = fx.Decorate(func() Probe { return NopProbe{} })

type NopProbe struct{}

func (p NopProbe) HabitCreated()             {}
func (p NopProbe) ActivityCreated()          {}
func (p NopProbe) FailedToCreateHabit(error) {}
