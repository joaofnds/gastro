package config

import (
	"os"

	"astro/metrics"
	"astro/token"

	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const path = "CONFIG_PATH"

var Module = fx.Options(
	fx.Invoke(LoadConfig),
	fx.Provide(NewAppConfig),
	fx.Provide(func(app App) HTTP { return app.HTTP }),
	fx.Provide(func(app App) Postgres { return app.Postgres }),
	fx.Provide(func(app App) token.Config { return app.Token }),
	fx.Provide(func(app App) metrics.Config { return app.Metrics }),
)

type App struct {
	Env      string         `mapstructure:"env"`
	HTTP     HTTP           `mapstructure:"http"`
	Postgres Postgres       `mapstructure:"postgres"`
	Token    token.Config   `mapstructure:"token"`
	Metrics  metrics.Config `mapstructure:"metrics"`
}

func init() {
	viper.MustBindEnv("env", "ENV")
	viper.MustBindEnv("metrics.address", "METRICS_ADDRESS")
	viper.MustBindEnv("token.public_key", "TOKEN_PUBLIC_KEY")
	viper.MustBindEnv("token.private_key", "TOKEN_PRIVATE_KEY")
}

func LoadConfig(logger *zap.Logger) error {
	configFile := os.Getenv(path)
	if configFile == "" {
		logger.Warn("could not lookup config path, will skip config file load")
		return nil
	}

	viper.SetConfigFile(configFile)
	return viper.ReadInConfig()
}

func NewAppConfig() (App, error) {
	var config App
	return config, viper.UnmarshalExact(&config)
}
