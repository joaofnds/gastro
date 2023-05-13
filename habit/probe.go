package habit

type Probe interface {
	FailedToCreateHabit(error)
	HabitCreated()
}
