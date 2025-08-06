package repo

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewUserRepo),
	fx.Provide(NewDeviceRepo),
	fx.Provide(NewCollectorRepo),
	fx.Provide(NewCollectTaskRepo),
	fx.Provide(NewDeviceTypeRepo),
	fx.Provide(NewDeviceLogRepo),
)
