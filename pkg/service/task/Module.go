package task

import (
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewCollectTaskServer),
	// 取消上报任务，改为采集后直接上报
	//fx.Provide(NewReportTaskServer),
)
