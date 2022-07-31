package test

import (
	"astro/habit"

	"go.uber.org/fx"
)

var FakeInstrumentation = fx.Decorate(NewFakeHabitInstrumentation)

type FakeHabitInstrumentation struct{}

func NewFakeHabitInstrumentation() habit.HabitInstrumentation {
	return &FakeHabitInstrumentation{}
}
func (l *FakeHabitInstrumentation) LogFailedToCreateHabit(err error) {}
func (l *FakeHabitInstrumentation) LogHabitCreated()                 {}
