package config

import (
	"astro/adapters/backuphealth"
	"astro/adapters/http"
	"astro/adapters/metrics"
	"astro/adapters/sqlite"
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
	fx.Provide(func(app App) sqlite.Config { return app.SQLite }),
	fx.Provide(func(app App) token.Config { return app.Token }),
	fx.Provide(func(app App) metrics.Config { return app.Metrics }),
	fx.Provide(func(app App) backuphealth.Config { return app.BackupHealth }),
)

type App struct {
	Env          string              `mapstructure:"env"`
	HTTP         http.Config         `mapstructure:"http"`
	SQLite       sqlite.Config       `mapstructure:"sqlite"`
	Token        token.Config        `mapstructure:"token"`
	Metrics      metrics.Config      `mapstructure:"metrics"`
	BackupHealth backuphealth.Config `mapstructure:"backup_health"`
}

func LoadConfig(logger *zap.Logger) error {
	configFile := os.Getenv(path)
	if configFile == "" {
		bindEnv()
		logger.Warn(path + " not set, will load config from env")
		return nil
	}

	viper.SetConfigFile(configFile)
	return viper.ReadInConfig()
}

func NewAppConfig() (App, error) {
	var config App
	return config, viper.UnmarshalExact(&config)
}

func bindEnv() {
	viper.MustBindEnv("env", "ENV")
	viper.MustBindEnv("metrics.address", "METRICS_ADDRESS")
	viper.MustBindEnv("http.port", "HTTP_PORT")
	viper.MustBindEnv("http.limiter.requests", "HTTP_LIMITER_REQUESTS")
	viper.MustBindEnv("http.limiter.expiration", "HTTP_LIMITER_EXPIRATION")
	viper.MustBindEnv("token.public_key", "TOKEN_PUBLIC_KEY")
	viper.MustBindEnv("token.private_key", "TOKEN_PRIVATE_KEY")
	viper.MustBindEnv("sqlite.path", "SQLITE_PATH")
	viper.MustBindEnv("backup_health.url", "BACKUP_HEALTHCHECK_URL")
}
