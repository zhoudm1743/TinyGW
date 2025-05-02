package collect

import (
	"go.uber.org/fx"
	"tinyGW/pkg/service/collect/channel"
)

var Module = fx.Options(
	channel.Module,
	fx.Provide(NewCollectorServer),
)
