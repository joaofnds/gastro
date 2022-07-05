package test

import (
	"astro/config"
	"math/rand"
)

func RandomAppConfigPort(config config.AppConfig) config.AppConfig {
	config.Port = 10_000 + rand.Intn(5000)
	return config
}
