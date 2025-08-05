package event

import "go.uber.org/fx"

var Module = fx.Provide(NewEventBus)
