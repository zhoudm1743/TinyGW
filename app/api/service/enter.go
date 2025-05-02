package service

import "go.uber.org/fx"

var Module = fx.Module("service",
	fx.Provide(NewAccountService),
	fx.Provide(NewCollectorService),
	fx.Provide(NewCollectTaskService),
	fx.Provide(NewReportTaskService),
	fx.Provide(NewDeviceService),
	fx.Provide(NewDeviceTypeService),
	fx.Provide(NewUserService),
)
