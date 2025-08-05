package repository

import "go.uber.org/fx"

var Module = fx.Module("repository",
	fx.Provide(
		NewCollectTaskRepository,
		NewCollectorRepository,
		NewUserRepository,
		NewDeviceRepository,
		NewDeviceTypeRepository,
		NewReportTaskRepository,
	),
)
