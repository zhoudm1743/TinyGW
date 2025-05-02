package task

import (
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewCollectTaskServer),
	fx.Provide(NewReportTaskServer),
)
