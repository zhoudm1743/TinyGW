package script

import "go.uber.org/fx"

var Module = fx.Provide(NewRunner)
