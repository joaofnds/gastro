package test

import (
	"astro/adapters/http"

	"go.uber.org/fx"
)

var (
	i                int
	preAllocPorts    = 1_000
	ports            = FindPorts(10_000, preAllocPorts)
	NewPortAppConfig = fx.Decorate(newPortAppConfig)
)

func newPortAppConfig(httpConfig http.Config) http.Config {
	httpConfig.Port = ports[i] // if this fail we ran out of ports, just increase `preAllocPorts`
	i++
	return httpConfig
}
