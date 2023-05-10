package token

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

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

func (i *PromProbe) TokenCreated() {
	i.logger.Info("token created")
	i.tokensCreated.Inc()
}

func (i *PromProbe) FailedToCreateToken(err error) {
	i.logger.Error("failed to decrypt token", zap.Error(err))
}

func (i *PromProbe) TokenDecrypted() {
	i.logger.Info("token decrypted")
	i.decrypts.Inc()
}

func (i *PromProbe) FailedToDecryptToken(err error) {
	i.logger.Error("failed to decrypt token", zap.Error(err))
	i.decryptsFailed.Inc()
}
