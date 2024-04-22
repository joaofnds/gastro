package habit

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var _ Probe = (*PromProbe)(nil)

type PromProbe struct {
	logger             *zap.Logger
	habitsCreated      prometheus.Counter
	habitsCreateFailed prometheus.Counter
	activityCreated    prometheus.Counter
}

func NewPromProbe(logger *zap.Logger) *PromProbe {
	return &PromProbe{
		logger:             logger,
		habitsCreated:      promauto.NewCounter(prometheus.CounterOpts{Name: "astro_habits_created"}),
		habitsCreateFailed: promauto.NewCounter(prometheus.CounterOpts{Name: "astro_habits_create_fail"}),
		activityCreated:    promauto.NewCounter(prometheus.CounterOpts{Name: "astro_activities_created"}),
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

func (p *PromProbe) ActivityCreated() {
	p.logger.Info("activity created")
	p.activityCreated.Inc()
}
