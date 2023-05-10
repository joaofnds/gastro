package habit

import (
	"go.uber.org/fx"
)

var NopProbeProvider = fx.Decorate(func() Probe {
	return &NopProbe{}
})

type NopProbe struct{}

func (l *NopProbe) LogFailedToCreateHabit(error) {}
func (l *NopProbe) LogHabitCreated()             {}
