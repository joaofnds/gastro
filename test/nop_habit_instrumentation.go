package test

import (
	"astro/habit"

	"go.uber.org/fx"
)

var NopHabitInstrumentation = fx.Decorate(NewNopHabitInstrumentation)

type nopHabitInstrumentation struct{}

func NewNopHabitInstrumentation() habit.HabitInstrumentation {
	return &nopHabitInstrumentation{}
}
func (l *nopHabitInstrumentation) LogFailedToCreateHabit(err error) {}
func (l *nopHabitInstrumentation) LogHabitCreated()                 {}
