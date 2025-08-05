package api

import (
	"go.uber.org/fx"
	"tinyGW/app/api/repository"
	"tinyGW/app/api/routes"
	"tinyGW/app/api/service"
)

var Module = fx.Options(
	repository.Module,
	service.Module,
	routes.Module,
)
