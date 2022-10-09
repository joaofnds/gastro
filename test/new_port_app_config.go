package test

import (
	"astro/config"

	"go.uber.org/fx"
)

var (
	i                int
	preAllocPorts    = 1_000
	ports            = FindPorts(10_000, preAllocPorts)
	NewPortAppConfig = fx.Decorate(newPortAppConfig)
)

func newPortAppConfig(config config.AppConfig) config.AppConfig {
	config.Port = ports[i] // if this fail we ran out of ports, just increase `preAllocPorts`
	i++
	return config
}
