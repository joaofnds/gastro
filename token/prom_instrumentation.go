package token

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

type PromTokenInstrumentation struct {
	logger         *zap.Logger
	tokensCreated  prometheus.Counter
	decrypts       prometheus.Counter
	decryptsFailed prometheus.Counter
}

func NewPromTokenInstrumentation(logger *zap.Logger) *PromTokenInstrumentation {
	return &PromTokenInstrumentation{
		logger:         logger,
		tokensCreated:  promauto.NewCounter(prometheus.CounterOpts{Name: "astro_token_created"}),
		decrypts:       promauto.NewCounter(prometheus.CounterOpts{Name: "astro_token_decrypts"}),
		decryptsFailed: promauto.NewCounter(prometheus.CounterOpts{Name: "astro_token_decrypts_failed"}),
	}
}

func (i *PromTokenInstrumentation) TokenCreated() {
	i.logger.Info("token created")
	i.tokensCreated.Inc()
}

func (i *PromTokenInstrumentation) FailedToCreateToken(err error) {
	i.logger.Error("failed to decrypt token", zap.Error(err))
}

func (i *PromTokenInstrumentation) TokenDecrypted() {
	i.logger.Info("token decrypted")
	i.decrypts.Inc()
}

func (i *PromTokenInstrumentation) FailedToDecryptToken(err error) {
	i.logger.Error("failed to decrypt token", zap.Error(err))
	i.decryptsFailed.Inc()
}
