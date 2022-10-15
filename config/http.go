package config

import (
	"github.com/spf13/viper"
	"time"
)

type HTTP struct {
	Port    int     `mapstructure:"port"`
	Limiter Limiter `mapstructure:"limiter"`
}

type Limiter struct {
	Requests   int           `mapstructure:"requests"`
	Expiration time.Duration `mapstructure:"expiration"`
}

func init() {
	viper.MustBindEnv("http.port", "HTTP_PORT")
	viper.MustBindEnv("http.limiter.requests", "HTTP_LIMITER_REQUESTS")
	viper.MustBindEnv("http.limiter.expiration", "HTTP_LIMITER_EXPIRATION")
}
