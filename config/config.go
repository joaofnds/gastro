package config

import (
	"astro/adapters/http"
	"astro/adapters/metrics"
	"astro/adapters/postgres"
	"os"

	"astro/token"

	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const path = "CONFIG_PATH"

var Module = fx.Module(
	"config",
	fx.Invoke(LoadConfig),
	fx.Provide(NewAppConfig),
	fx.Provide(func(app App) http.Config { return app.HTTP }),
	fx.Provide(func(app App) postgres.Config { return app.Postgres }),
	fx.Provide(func(app App) token.Config { return app.Token }),
	fx.Provide(func(app App) metrics.Config { return app.Metrics }),
)

type App struct {
	Env      string          `mapstructure:"env"`
	HTTP     http.Config     `mapstructure:"http"`
	Postgres postgres.Config `mapstructure:"postgres"`
	Token    token.Config    `mapstructure:"token"`
	Metrics  metrics.Config  `mapstructure:"metrics"`
}

func init() {
	viper.MustBindEnv("env", "ENV")
	viper.MustBindEnv("metrics.address", "METRICS_ADDRESS")
	viper.MustBindEnv("http.port", "HTTP_PORT")
	viper.MustBindEnv("http.limiter.requests", "HTTP_LIMITER_REQUESTS")
	viper.MustBindEnv("http.limiter.expiration", "HTTP_LIMITER_EXPIRATION")
	viper.MustBindEnv("token.public_key", "TOKEN_PUBLIC_KEY")
	viper.MustBindEnv("token.private_key", "TOKEN_PRIVATE_KEY")
	viper.MustBindEnv("postgres.host", "POSTGRES_HOST")
	viper.MustBindEnv("postgres.port", "POSTGRES_PORT")
	viper.MustBindEnv("postgres.user", "POSTGRES_USER")
	viper.MustBindEnv("postgres.password", "POSTGRES_PASSWORD")
	viper.MustBindEnv("postgres.dbname", "POSTGRES_DBNAME")
}

func LoadConfig(logger *zap.Logger) error {
	configFile := os.Getenv(path)
	if configFile == "" {
		logger.Warn("could not lookup config path, skipping config file load")
		return nil
	}

	viper.SetConfigFile(configFile)
	return viper.ReadInConfig()
}

func NewAppConfig() (App, error) {
	var config App
	return config, viper.UnmarshalExact(&config)
}
