package boot

import (
	"TinyGW/core/cache"
	"TinyGW/core/config"
	"TinyGW/core/database"
	"TinyGW/core/event"
	"TinyGW/core/logger"
	"TinyGW/core/script"
	"TinyGW/core/web"
	"TinyGW/internal"

	"go.uber.org/fx"
)

var Module = fx.Options(
	config.Module,
	database.Module,
	event.Module,
	logger.Module,
	cache.Module,
	web.Module,
	script.Module,
	internal.Module,
)
