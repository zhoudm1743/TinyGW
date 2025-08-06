package controller

import (
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		NewUserController,
		NewDeviceController,
		NewCollectorController,
		NewCollectTaskController,
		NewDeviceTypeController,
		NewDeviceLogController,
		NewAuthController,
	),
)
