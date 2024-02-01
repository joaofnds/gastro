package token

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var _ Probe = (*PromProbe)(nil)

type PromProbe struct {
	logger         *zap.Logger
	tokensCreated  prometheus.Counter
	decrypts       prometheus.Counter
	decryptsFailed prometheus.Counter
}

func NewPromProbe(logger *zap.Logger) *PromProbe {
	return &PromProbe{
		logger:         logger,
		tokensCreated:  promauto.NewCounter(prometheus.CounterOpts{Name: "astro_token_created"}),
		decrypts:       promauto.NewCounter(prometheus.CounterOpts{Name: "astro_token_decrypts"}),
		decryptsFailed: promauto.NewCounter(prometheus.CounterOpts{Name: "astro_token_decrypts_failed"}),
	}
}

func (p *PromProbe) TokenCreated() {
	p.logger.Info("token created")
	p.tokensCreated.Inc()
}

func (p *PromProbe) FailedToCreateToken(err error) {
	p.logger.Error("failed to decrypt token", zap.Error(err))
}

func (p *PromProbe) TokenDecrypted() {
	p.logger.Info("token decrypted")
	p.decrypts.Inc()
}

func (p *PromProbe) FailedToDecryptToken(err error) {
	p.logger.Error("failed to decrypt token", zap.Error(err))
	p.decryptsFailed.Inc()
}
