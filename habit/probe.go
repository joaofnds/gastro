package habit

type Probe interface {
	LogFailedToCreateHabit(error)
	LogHabitCreated()
}
