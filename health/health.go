package health

import "go.uber.org/fx"

var Module = fx.Provide(NewHealthService)
