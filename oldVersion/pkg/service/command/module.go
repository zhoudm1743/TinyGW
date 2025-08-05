package command

import (
	"go.uber.org/fx"
)

// Module 命令管理模块
var Module = fx.Options(
	fx.Provide(GetManager),
)
