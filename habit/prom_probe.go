package habit

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

type PromProbe struct {
	logger             *zap.Logger
	habitsCreated      prometheus.Counter
	habitsCreateFailed prometheus.Counter
}

func NewPromProbe(logger *zap.Logger) *PromProbe {
	return &PromProbe{
		logger:             logger,
		habitsCreated:      promauto.NewCounter(prometheus.CounterOpts{Name: "astro_habits_created"}),
		habitsCreateFailed: promauto.NewCounter(prometheus.CounterOpts{Name: "astro_habits_create_fail"}),
	}
}

func (p *PromProbe) FailedToCreateHabit(err error) {
	p.logger.Info("failed to create habit", zap.Error(err))
	p.habitsCreateFailed.Inc()
}

func (p *PromProbe) HabitCreated() {
	p.logger.Info("habit created")
	p.habitsCreated.Inc()
}
