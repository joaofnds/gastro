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

func NewPromHabitInstrumentation(logger *zap.Logger) *PromHabitInstrumentation {
	return &PromHabitInstrumentation{
		logger:             logger,
		habitsCreated:      promauto.NewCounter(prometheus.CounterOpts{Name: "astro_habits_created"}),
		habitsCreateFailed: promauto.NewCounter(prometheus.CounterOpts{Name: "astro_habits_create_fail"}),
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
