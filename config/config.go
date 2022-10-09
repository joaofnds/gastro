package config

import (
	"os"

	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const configPath = "CONFIG_PATH"

var Module = fx.Options(
	fx.Invoke(LoadConfig),
	fx.Provide(NewAppConfig),
	fx.Provide(NewPostgresConfig),
	fx.Provide(NewTokenConfig),
	fx.Provide(NewMetricsConfig),
)

type AppConfig struct {
	Env      string         `mapstructure:"env"`
	Port     int            `mapstructure:"port"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Token    TokenConfig    `mapstructure:"token"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
}

func init() {
	viper.MustBindEnv("env", "ENV")
	viper.MustBindEnv("port", "PORT")
}

func LoadConfig(logger *zap.Logger) error {
	configFile, ok := os.LookupEnv(configPath)
	if !ok || configFile == "" {
		logger.Warn("could not lookup config path, will skip config file load")
		return nil
	}

	viper.SetConfigFile(configFile)
	return viper.ReadInConfig()
}

func NewAppConfig() (AppConfig, error) {
	var config AppConfig

	if err := viper.UnmarshalExact(&config); err != nil {
		return config, err
	}

	return config, nil
}

func NewPostgresConfig(appConfig AppConfig) PostgresConfig { return appConfig.Postgres }
func NewTokenConfig(appConfig AppConfig) TokenConfig       { return appConfig.Token }
func NewMetricsConfig(appConfig AppConfig) MetricsConfig   { return appConfig.Metrics }
