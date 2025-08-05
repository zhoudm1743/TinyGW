package cache

import "go.uber.org/fx"

var Module = fx.Provide(NewCache)
