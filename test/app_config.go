package test

import (
	"astro/config"
)

var (
	i             int
	preAllocPorts = 1_000
	ports         = FindPorts(10_000, preAllocPorts)
)

func RandomAppConfigPort(config config.AppConfig) config.AppConfig {
	config.Port = ports[i] // if this fail we ran out of ports, just increase `preAllocPorts`
	i++
	return config
}
