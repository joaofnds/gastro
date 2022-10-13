package habit

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

type Instrumentation interface {
	LogFailedToCreateHabit(error)
	LogHabitCreated()
}

type PromInstrumentation struct {
	logger             *zap.Logger
	habitsCreated      prometheus.Counter
	habitsCreateFailed prometheus.Counter
}

func NewPromInstrumentation(logger *zap.Logger) *PromInstrumentation {
	return &PromInstrumentation{
		logger:             logger,
		habitsCreated:      promauto.NewCounter(prometheus.CounterOpts{Name: "astro_habits_created"}),
		habitsCreateFailed: promauto.NewCounter(prometheus.CounterOpts{Name: "astro_habits_create_fail"}),
	}
}

func (l *PromInstrumentation) LogFailedToCreateHabit(err error) {
	l.logger.Info("failed to create habit", zap.Error(err))
	l.habitsCreateFailed.Inc()
}

func (l *PromInstrumentation) LogHabitCreated() {
	l.logger.Info("habit created")
	l.habitsCreated.Inc()
}
