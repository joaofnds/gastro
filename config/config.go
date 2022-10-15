package config

import (
	"os"

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
	fx.Provide(func(app App) Token { return app.Token }),
	fx.Provide(func(app App) Metrics { return app.Metrics }),
)

type App struct {
	Env      string   `mapstructure:"env"`
	HTTP     HTTP     `mapstructure:"http"`
	Postgres Postgres `mapstructure:"postgres"`
	Token    Token    `mapstructure:"token"`
	Metrics  Metrics  `mapstructure:"metrics"`
}

func init() {
	viper.MustBindEnv("env", "ENV")
}

func LoadConfig(logger *zap.Logger) error {
	configFile, ok := os.LookupEnv(path)
	if !ok || configFile == "" {
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
