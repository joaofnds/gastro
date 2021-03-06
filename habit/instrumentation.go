package habit

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

type HabitInstrumentation interface {
	LogFailedToCreateHabit(error)
	LogHabitCreated()
}

type PromHabitInstrumentation struct {
	logger             *zap.Logger
	habitsCreated      prometheus.Counter
	habitsCreateFailed prometheus.Counter
}

func NewPromHabitInstrumentation(logger *zap.Logger) HabitInstrumentation {
	return &PromHabitInstrumentation{
		logger:             logger,
		habitsCreated:      promauto.NewCounter(prometheus.CounterOpts{Name: "habits_created"}),
		habitsCreateFailed: promauto.NewCounter(prometheus.CounterOpts{Name: "habits_create_fail"}),
	}
}

func (l *PromHabitInstrumentation) LogFailedToCreateHabit(err error) {
	l.logger.Info("failed to create habit", zap.Error(err))
	l.habitsCreateFailed.Inc()
}

func (l *PromHabitInstrumentation) LogHabitCreated() {
	l.logger.Info("habit created")
	l.habitsCreated.Inc()
}
