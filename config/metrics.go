package config

import (
	"github.com/spf13/viper"
)

type Metrics struct {
	Address string `mapstructure:"address"`
}

func init() {
	viper.MustBindEnv("metrics.address", "METRICS_ADDRESS")
}
